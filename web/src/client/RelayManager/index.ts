import { TextMessage } from "@Src/storage/message";
import BaseEventSystem from "@Src/structs/eventSystem";
import Authorization from "@API/Authorization";
import MsgBridgeAPI, { LinkCode } from "@Src/api/Relay";

type ConnectionState = "INITIAL" | "UNCONNECTED" | "CONNECTING" | "CONNECTED";

export interface UserData {
    userId: string;
    channelId: string;
    publicKey?: CryptoKey;
}

export class UserData {
    constructor(userId: string, publicKey?: CryptoKey) {
        this.userId = userId;
        if (publicKey) {
            this.publicKey = publicKey;
        }
    }

    public SetChannelId(channelId: string): void {
        this.channelId = channelId;
    }
}

export interface MessageData {
    sender: UserData;
    message: TextMessage;
}

type EventDefinitions = {
    "StateChanged": { detail: ConnectionState };
    "MessageAppended": { detail: MessageData };
    "UserAppended": { detail: UserData };
    "LinkCodeDurationUpdated": { detail: number };
    "ErrorOccurred": { detail: { message: string, error: Error } };
};

export type RelayManagerEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

export default class RelayManager extends BaseEventSystem<EventDefinitions> {
    public static Status = class {
        public static readonly INITIAL: ConnectionState = "INITIAL";
        public static readonly UNCONNECTED: ConnectionState = "UNCONNECTED";
        public static readonly CONNECTING: ConnectionState = "CONNECTING";
        public static readonly CONNECTED: ConnectionState = "CONNECTED";
    };

    private connectionState: ConnectionState = RelayManager.Status.INITIAL;
    private api?: MsgBridgeAPI;
    private users: Map<string, UserData>;
    private linkCode?: LinkCode;

    constructor(auth: Authorization) {
        super();

        const users = new Map<string, UserData>();

        MsgBridgeAPI.GenerateKey().then((key) => {
            const api = new MsgBridgeAPI(auth, key);

            api.on("ConnectedChanged", (e) => {
                if (e.detail) {
                    this.SetState(RelayManager.Status.CONNECTED);
                } else {
                    this.SetState(RelayManager.Status.UNCONNECTED);
                }
            });

            api.on("UserAppended", (e) => {
                const { clientId, publicKey } = e.detail;
                const user = new UserData(clientId, publicKey);
                users.set(clientId, user);
                this.emit("UserAppended", { detail: user });
            });

            api.on("MessageReceived", (e) => {
                const { sender_id, message } = e.detail;
                let user = users.get(sender_id);

                if (!user) {
                    user = new UserData(sender_id);
                    users.set(sender_id, user);
                }

                this.emit("MessageAppended", {
                    detail: {
                        sender: user,
                        message: TextMessage.New(message),
                    },
                });
            });

            api.on("ErrorOccurred", (e) => {
                this.emit("ErrorOccurred", e);
            });

            this.api = api;
        });

        this.users = users;

        setInterval(() => {
            const { linkCode } = this;
            if (linkCode) {
                const duration = linkCode.expires_at - Math.floor(Date.now() / 1000);
                if (duration <= 0) {
                    this.linkCode = undefined;
                }

                this.emit("LinkCodeDurationUpdated", { detail: duration });
            }
        }, 1000);
    }

    private SetState(state: ConnectionState): void {
        this.connectionState = state;
        this.emit("StateChanged", { detail: state });
    }

    public GetState(): ConnectionState {
        return this.connectionState;
    }

    public async RequestLinkCode(): Promise<string> {
        const { api } = this;

        return new Promise(async (resolve, reject) => {
            if (!api) {
                reject(new Error("API not available"));
                return;
            }

            api.UploadKey()
                .then((linkCode) => {
                    this.linkCode = linkCode;
                    resolve(linkCode.link_code);
                })
                .catch(reject);
        });
    }

    public async RemoveLinkCode(): Promise<void> {
        const { api, linkCode } = this;

        return new Promise((resolve, reject) => {
            if (!api) {
                reject(new Error("API not available"));
                return;
            }

            if (!linkCode) {
                reject(new Error("No link code to remove"));
                return;
            }

            api.RemoveKey(linkCode.link_code).then(resolve).catch(reject);
            linkCode.expires_at = 0;
        });
    }

    public async GetUserSession(linkCode: string): Promise<void> {
        const { api, users } = this;

        if (!api) {
            throw new Error("API not available");
        }

        const session = await api.GetUserSession(linkCode);

        const user = new UserData(session.client_id, session.public_key);
        users.set(session.client_id, user);
        this.emit("UserAppended", { detail: user });
    }

    public async SendMessage(targetId: string, content: string): Promise<TextMessage> {
        const { api, users } = this;

        if (!api) {
            throw new Error("API not available");
        }

        const user = users.get(targetId);

        if (!user) {
            throw new Error("User not found");
        }

        if (!user.publicKey) {
            throw new Error("User does not have a public key");
        }

        await api.SendMessage(targetId, user.publicKey, content);

        return TextMessage.New(content);
    }

    public Enable(): Promise<void> {
        const { api } = this;

        if (!api) {
            throw new Error("API not available");
        }

        this.SetState(RelayManager.Status.CONNECTING);
        return api.Connect();
    }

    public Disable(): void {
        const { api } = this;
        if (api && api.State === MsgBridgeAPI.STATUS.READY) {
            api.CloseConnection();
            this.SetState(RelayManager.Status.UNCONNECTED);
        }
    }

    public get Enabled(): boolean {
        const { api } = this;
        return api !== undefined && api.State === MsgBridgeAPI.STATUS.READY;
    }

    public get LinkCode(): LinkCode | undefined {
        return this.linkCode;
    }
}