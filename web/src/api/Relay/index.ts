import SseParser, { MessageData, SessionData } from "./sseParser";
import BaseEventSystem from "@Src/structs/eventSystem";
import Authorization from "@API/Authorization";

export interface LinkCode {
    link_code: string;
    expires_at: number;
}

export interface UserData {
    clientId: string;
    publicKey: CryptoKey;
}

export type {
    MessageData,
    SessionData,
}

type State = "INITIAL" | "READY" | "CLOSED" | "ERROR";

type EventDefinitions = {
    "StateChanged": { detail: State };
    "ConnectedChanged": { detail: boolean };
    "MessageReceived": { detail: MessageData };
    "UserAppended": { detail: UserData };
    "ErrorOccurred": { detail: { message: string, error: Error } };
};

export type MessageBridgeEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

export default class MessageBridgeApi extends BaseEventSystem<EventDefinitions> {
    public static KEY_NAME = "RSA-OAEP";
    public static KEY_FORMAT: Exclude<KeyFormat, "jwk"> = "spki";
    public static HASH_NAME = "SHA-256";

    // Constant URLs
    private static API_BASE_URL = "./api/message";
    private static MESSAGES_URL = `${MessageBridgeApi.API_BASE_URL}/messages`;
    private static SESSION_URL = `${MessageBridgeApi.API_BASE_URL}/session`;

    public static GenerateKey = (): Promise<CryptoKeyPair> => {
        return crypto.subtle.generateKey(
            {
                name: MessageBridgeApi.KEY_NAME,
                modulusLength: 2048,
                publicExponent: new Uint8Array([1, 0, 1]),
                hash: {
                    name: MessageBridgeApi.HASH_NAME,
                },
            },
            true,
            ["encrypt", "decrypt"],
        );
    };

    public static STATUS = class {
        public static INITIAL: State = "INITIAL";
        public static READY: State = "READY";
        public static CLOSED: State = "CLOSED";
        public static ERROR: State = "ERROR";
    };

    private state: State = MessageBridgeApi.STATUS.INITIAL;
    private controller?: AbortController;
    private parser: SseParser;
    private auth: Authorization;
    private key: CryptoKeyPair;

    constructor(auth: Authorization, key: CryptoKeyPair) {
        super();
        this.auth = auth;
        this.key = key;
        this.parser = new SseParser({
            key_name: MessageBridgeApi.KEY_NAME,
            key_format: MessageBridgeApi.KEY_FORMAT,
            hash_name: MessageBridgeApi.HASH_NAME,
            private_key: key.privateKey,
        });
    }

    public async Connect(): Promise<void> {
        const { parser, state, auth } = this;

        try {
            if (state === MessageBridgeApi.STATUS.READY) {
                return Promise.reject(new Error("Already connected"));
            }

            const controller = new AbortController();

            const response = await fetch(MessageBridgeApi.MESSAGES_URL, {
                method: "GET",
                headers: {
                    "Authorization": `Bearer ${auth.AccessToken}`,
                },
                signal: controller.signal,
            });

            if (!response.body) {
                return Promise.reject(new Error("Response body is empty"));
            }

            if (!response.ok) {
                return Promise.reject(new Error(await response.text()));
            }

            this.SetState(MessageBridgeApi.STATUS.READY);
            this.SetConnected(controller);

            const reader = response.body.getReader();

            while (!controller.signal.aborted) {
                const { done, value } = await reader.read();
                if (done) {
                    this.CloseConnection();
                    return;
                }

                try {
                    const { event, data } = SseParser.ParseResponse(value);

                    switch (event) {
                        case "connected":
                            break;
                        case "append_user":
                            const sessionData = await parser.DecodeSessionData(data);
                            this.emit("UserAppended", {
                                detail: {
                                    clientId: sessionData.client_id,
                                    publicKey: sessionData.public_key,
                                },
                            });
                            break;
                        case "message":
                            const messageData = await parser.DecodeMessageData(data);
                            this.emit("MessageReceived", {
                                detail: messageData,
                            });
                            break;
                    }

                } catch (decryptError) {
                    this.HandleError("Decryption or decompression error", decryptError);
                }
            }

        } catch (error) {
            this.HandleError("Initialization error", error);
            this.SetState(MessageBridgeApi.STATUS.ERROR);
        }
    }

    private HandleError(message: string, rawError: unknown): void {
        let error: Error;

        if (rawError instanceof Error) {
            if (rawError.name === 'AbortError') {
                return;
            }
            error = rawError;
        } else if (typeof rawError === "string") {
            error = new Error(rawError);
        } else if (typeof rawError === "object") {
            error = new Error(JSON.stringify(rawError));
        } else {
            error = new Error(String(rawError));
        }

        this.emit("ErrorOccurred", { detail: { message, error } });
    }

    private SetState(state: State): void {
        this.state = state;
        this.emit("StateChanged", { detail: state });
    }

    public get State(): State {
        return this.state;
    }

    private SetConnected(controller?: AbortController): void {
        if (!controller && this.controller) {
            this.controller.abort();
        }

        this.controller = controller;
        this.emit("ConnectedChanged", { detail: controller !== undefined });
    }

    public get IsConnected(): boolean {
        return this.controller !== undefined;
    }

    public UploadKey(): Promise<LinkCode> {
        const { state, auth, key, parser } = this;

        return new Promise(async (resolve, reject) => {
            if (state !== MessageBridgeApi.STATUS.READY) {
                reject(new Error("Cannot set key: The system is not ready"));
                return;
            }

            try {
                const response = await fetch(MessageBridgeApi.SESSION_URL, {
                    method: "POST",
                    headers: {
                        "Authorization": `Bearer ${auth.AccessToken}`,
                    },
                    body: await parser.EncodeSessionData(key.publicKey),
                });

                if (!response.ok) {
                    throw new Error(`Failed to set key, status: ${response.status}`);
                }

                const result: LinkCode = await response.json();
                resolve(result);

            } catch (error) {
                reject(error);
            }
        });
    }

    public async RemoveKey(linkCode: string): Promise<void> {
        const { auth } = this;

        await fetch(`${MessageBridgeApi.SESSION_URL}/${linkCode}`, {
            method: "DELETE",
            headers: {
                "Authorization": `Bearer ${auth.AccessToken}`,
            },
        });
    }

    public async GetUserSession(linkCode: string): Promise<SessionData> {
        const { parser, key } = this;

        const response = await fetch(`${MessageBridgeApi.SESSION_URL}/${linkCode}`, {
            method: "POST",
            headers: {
                "Authorization": `Bearer ${this.auth.AccessToken}`,
            },
            body: await parser.EncodeSessionData(key.publicKey),
        });

        const rawSessionData = await response.text();
        return parser.DecodeSessionData(rawSessionData);
    }

    public async SendMessage(targetId: string, publicKey: CryptoKey, content: string): Promise<void> {
        try {
            const data = await this.parser.EncodeMessageData(publicKey, content);

            await fetch(`${MessageBridgeApi.MESSAGES_URL}/${targetId}`, {
                method: "POST",
                headers: { "Authorization": `Bearer ${this.auth.AccessToken}` },
                body: data,
            });
        } catch (error) {
            this.HandleError("Sending message failed", error);
        }
    }

    public CloseConnection(): void {
        this.SetConnected();
        this.SetState(MessageBridgeApi.STATUS.CLOSED);
    }
}
