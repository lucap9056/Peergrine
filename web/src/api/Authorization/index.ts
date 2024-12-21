import BaseEventSystem from "@Src/structs/eventSystem";
import JWT, { Payload } from "./jwt";


export type Message<T> = {
    type: string;
    content: T;
};

type BearerToken = {
    "access_token": string;
    "expires_at": number;
};

type RefreshToken = {
    "refresh_token": string;
} & BearerToken;

type EventDefinitions = {
    "MessageReceived": { detail: Message<any> }
    "AuthorizationStateChanged": { detail: Payload };
    "ErrorOccurred": { detail: { message: string, error: Error } };
};
export type AuthorizationEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

export default class Authorization extends BaseEventSystem<EventDefinitions> {

    private jwtInstance?: JWT;

    private static readonly WS_URL = "./api/token/initialize";
    private static readonly REFRESH_URL = "./api/token/refresh";

    private refreshTokenTimeout?: NodeJS.Timeout;

    constructor() {
        super();
        this.InitializeWebSocketConnection();
    }

    private InitializeWebSocketConnection() {
        const wss = new WebSocket(Authorization.WS_URL);

        wss.addEventListener("open", () => this.WebSocketOpenHandler());
        wss.addEventListener("close", () => this.WebSocketCloseHandler());
        wss.addEventListener("error", () => this.WebSocketErrorHandler());
        wss.addEventListener("message", (e) => this.WebSocketMessageHandler(e));

        setInterval(() => {
            if (wss.OPEN) {
                wss.send("PING")
            }
        }, 60 * 1000);
    }

    private WebSocketOpenHandler() {
        console.log("WebSocket connection established.");
    }

    private WebSocketMessageHandler(e: MessageEvent) {
        const originalMessage: Message<any> = JSON.parse(e.data.toString());

        if (originalMessage.type === "Authorization") {
            const message: Message<RefreshToken> = originalMessage;

            this.ProcessAuthorization(message.content);
            return;
        }
        
        this.emit("MessageReceived", { detail: originalMessage });
    }

    private WebSocketCloseHandler() {
        console.log("WebSocket connection closed.");
    }

    private WebSocketErrorHandler() {
        console.error("WebSocket error occurred.");
    }

    private ProcessAuthorization(refreshToken: RefreshToken) {
        const jwt = new JWT(refreshToken.access_token);
        this.jwtInstance = jwt;
        this.RefreshAuthToken(refreshToken);
        this.emit("AuthorizationStateChanged", { detail: jwt.Payload });
    }

    private RefreshAuthToken(auth: RefreshToken) {
        const { refreshTokenTimeout } = this;
        const { expires_at } = auth;
        const duration = expires_at * 1000 - new Date().getTime();


        if (refreshTokenTimeout) {
            clearTimeout(refreshTokenTimeout);
        }

        this.refreshTokenTimeout = setTimeout(async () => {

            try {
                const response = await fetch(Authorization.REFRESH_URL, {
                    method: "POST",
                    headers: {
                        "Authorization": auth.refresh_token
                    }
                });

                if (response.status !== 200) {
                    const msg = await response.text();
                    throw new Error(`Failed to refresh token: ${msg}`);
                }

                const newTokens: BearerToken = await response.json();
                this.jwtInstance = new JWT(newTokens.access_token);
                Object.assign(auth, newTokens);

                this.RefreshAuthToken(auth);

            } catch (error) {
                this.HandleError("Error while refreshing token:", error);
            }
        }, duration);
    }

    private HandleError(message: string, rawError: unknown): void {
        let error: Error;

        if (rawError instanceof Error) {
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

    public get Payload(): Payload | undefined {
        return this.jwtInstance?.Payload;
    }

    public get AccessToken(): string {
        return this.jwtInstance?.Token || "";
    }
}
