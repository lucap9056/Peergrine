import { Message, FileMessage, ReceivedMessage, TextMessage, FileRequestMessage, FileChunkMessage } from "@Src/storage/message";
import { FileChunk, FileInfo } from "@Src/storage/message/file";
import BaseEventSystem from "@Src/structs/eventSystem";
import Signaling, { LinkCode, SignalingEvent } from "@API/Signaling";
import ClientProfile, { ClientProfileEvent } from "@Src/client/ClientProfile";
import Conn from "./Conn";

export {
    Conn as RTCConn
}

type EventDefinitions = {
    "UserAppended": { detail: Conn }
    "OfferReady": { detail: { id: string, linkCode: LinkCode } }
    "ErrorOccurred": { detail: { conn: Conn, error: Error } }
    "LinkCodeDurationUpdate": { detail: number }
    "UserStatusChanged": { detail: Conn }
    "FocusUserChanged": { detail: Conn }

    "MessageAppended": { detail: { user: Conn, message: Message } }
}

export type RTCEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

export default class RTCManager extends BaseEventSystem<EventDefinitions> {

    private client: ClientProfile;
    private api: Signaling;

    private users: { [userId: string]: Conn } = {};
    private unconnectedOffers?: Conn;

    private linkCode?: LinkCode;

    private linkCodeDurationCountdown: NodeJS.Timeout = setInterval(() => {
        const { linkCode } = this;
        if (linkCode) {
            const duration = linkCode.expires_at - Math.floor(Date.now() / 1000);
            if (duration <= 0) {
                this.linkCode = undefined;
            }

            this.emit("LinkCodeDurationUpdate", { detail: duration });
        }
    }, 1000);

    constructor(client: ClientProfile, signaling: Signaling) {
        super();
        this.client = client;
        this.api = signaling;

        signaling.on("SignalReceived", (e: SignalingEvent<"SignalReceived">) => {
            const conn = this.unconnectedOffers;

            if (conn) {
                conn.SetConnectionTarget(e.detail);
            }
        });

        client.on("ClientNameChanged", (e: ClientProfileEvent<"ClientNameChanged">) => {
            this.ClientNameChange(e.detail);
        });

        this.on("UserAppended", (e) => {
            const user = e.detail;
            this.users[user.targetId] = user;

            user.on("UserStatusChanged", () => {
                this.emit("UserStatusChanged", { detail: user });
            });

            user.on("MessageAppended", (e) => {
                this.emit("MessageAppended", {
                    detail: {
                        user,
                        message: e.detail.message
                    }
                });
            });
        });

    }

    private ClientNameChange(name: string): void {
        Object.values(this.users).forEach(user => {
            user.Send(Conn.CHANNELS.STATE, {
                type: "CHANGE_USER_NAME",
                data: name
            });
        });
    }

    public async Offer(): Promise<void> {
        const { ClientId, ClientName } = this.client;
        const { api } = this;

        if (!api) {
            throw new Error("Signaling API instance is unavailable. Ensure that the signaling service is initialized.");
        }

        const user = new Conn(ClientId, ClientName);
        const connId = user.ConnId;

        user.on("Ready", (e) => {
            if (this.linkCode) {
                this.RemoveLinkCode();
            }
            this.emit("UserAppended", e);
        });

        user.on("Close", (e) => this.emit("UserStatusChanged", e));

        user.on("ErrorOccurred", (e) => {
            console.error(e);
            this.emit("ErrorOccurred", e);
        });

        user.on("SignalChanged", async (e) => {

            api.SetSignal(e.detail,
                (signal) => {
                    delete this.unconnectedOffers;
                    user.SetConnectionTarget(signal);
                }
            ).then((linkCode) => {

                this.emit("OfferReady", {
                    detail: {
                        id: connId,
                        linkCode
                    }
                });
                this.linkCode = linkCode;

            }).catch((err) => {
                this.emit("ErrorOccurred", { detail: { conn: user, error: err as Error } });
            });


        });
        await user.CreateOffer();
        this.unconnectedOffers = user;
    }

    public async Answer(linkCode: string): Promise<Conn> {
        const { ClientId, ClientName } = this.client;
        const { api } = this;

        return new Promise(async (resolve, reject) => {

            if (!api) {
                throw new Error("Signaling API instance is unavailable. Ensure that the signaling service is initialized.");
            }

            const targetSignal = await api.GetSignal(linkCode);

            const user = new Conn(ClientId, ClientName);

            user.on("Ready", (e) => {
                this.emit("UserAppended", e);
                resolve(user);
            });

            user.on("ErrorOccurred", (e) => this.emit("ErrorOccurred", e));

            user.on("Close", (e) => this.emit("UserStatusChanged", e));

            user.on("SignalChanged", (e) => {
                const signal = e.detail;

                api.ForwardSignal(linkCode, signal).catch(reject);
            });

            user.CreateAnswer(targetSignal);

        });
    }

    public get LinkCode(): LinkCode | undefined {
        return this.linkCode;
    }

    public RemoveLinkCode(): void {
        const { api, linkCode } = this;

        if (!api) {
            throw new Error("Cannot remove link code: Signaling API is not available.");
        }

        if (!linkCode) {
            throw new Error("Cannot remove link code: No active link code exists.");
        }

        linkCode.expires_at = 0;

        return api.RemoveSignal();
    }

    public SendTextMessage(userId: string, content: string): TextMessage {
        const user = this.users[userId];
        if (!user) {
            throw new Error(`User with ID ${userId} not found. Unable to send text message.`);
        }

        const message = TextMessage.New(content);
        user.Send(Conn.CHANNELS.MESSAGE, message);

        return message;
    }

    public SendFileMessage(userId: string, content: FileInfo): FileMessage {
        const user = this.users[userId];
        if (!user) {
            throw new Error(`User with ID ${userId} not found. Unable to send file message.`);
        }

        const message = FileMessage.New(content);
        user.Send(Conn.CHANNELS.MESSAGE, message);

        return message;
    }

    public SendFileChunkMessage(userId: string, content: FileChunk): FileChunkMessage {
        const user = this.users[userId];
        if (!user) {
            throw new Error(`User with ID ${userId} not found. Unable to send file chunk message.`);
        }

        const message = FileChunkMessage.New(content);
        user.Send(Conn.CHANNELS.MESSAGE, message);

        return message;
    }

    public SendFileRequestMessage(userId: string, fileId: string, index?: number): void {
        const user = this.users[userId];

        if (user) {
            const message = FileRequestMessage.New(fileId, index);

            user.Send(Conn.CHANNELS.MESSAGE, message);
        }
    }

    public MessageReceived(userId: string, messageId: string) {
        const user = this.users[userId];

        if (!user) {
            return;
        }

        const message = ReceivedMessage.New(messageId);
        user.Send(Conn.CHANNELS.MESSAGE, message);
    }

    public Close(): void {
        Object.values(this.users).forEach(user => {
            user.Close();
        });

        clearInterval(this.linkCodeDurationCountdown);
    }
}
