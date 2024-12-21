import Authorization, { Message, AuthorizationEvent } from "@API/Authorization";
import BaseEventSystem from "@Src/structs/eventSystem";

export type Candidate = {
    candidate: string;
    sdpMLineIndex: number;
    sdpMid: string;
};

export type LinkCode = {
    link_code: string;
    expires_at: number;
};

export type Signal = {
    client_id: string;
    channel_id: number;
    sdp: string;
    candidates: RTCIceCandidate[];
};

type EventDefinitions = {
    "SignalReceived": { detail: Signal }
    "ErrorOccurred": { detail: { message: string, error: Error } };
};
export type SignalingEvent<T extends keyof EventDefinitions> = EventDefinitions[T];
export default class Signaling extends BaseEventSystem<EventDefinitions> {
    private auth: Authorization;

    // URL constants
    private static readonly API_BASE_URL = "./api/signal/";
    private static readonly SIGNAL_URL = `${Signaling.API_BASE_URL}`;
    private static readonly SIGNAL_WITH_LINKCODE_URL = `${Signaling.API_BASE_URL}`;

    private setSignalController?: AbortController;

    constructor(authorization: Authorization) {
        super();
        this.auth = authorization;

        authorization.on("MessageReceived", (e: AuthorizationEvent<"MessageReceived">) => {
            if (e.detail.type !== "message-relay") return;

            const message: Message<Signal> = e.detail;

            this.emit("SignalReceived", { detail: message.content });
        });
    }

    /**
     * Sends a signaling request and handles the response.
     * 
     * @param signal - The signaling data to send.
     * @param Code - Callback to handle link code.
     * @param Target - Callback to handle SDP signals.
     */
    public async SetSignal(
        signal: Signal,
        Target: (signal: Signal) => void
    ): Promise<LinkCode> {
        const { auth } = this;

        const controller = new AbortController();
        this.setSignalController = controller;

        return new Promise(async (resolve, reject) => {

            const res = await fetch(Signaling.SIGNAL_URL, {
                method: "POST",
                body: JSON.stringify(signal),
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${auth.AccessToken}`,
                },
                signal: controller.signal
            });

            if (!res.body) {
                reject(new Error("Response body is empty"));
                return;
            }
            if (!res.ok) {
                reject(await res.text());
                return;
            }

            const reader = res.body.getReader();

            const decoder = new TextDecoder();
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const resultStr = decoder.decode(value);
                const result = JSON.parse(resultStr);
                console.log(result);
                if (result.link_code) {
                    resolve(result);
                }

                if (result.sdp) {
                    Target(result);
                }
            }

        });
    }

    /**
     * Retrieves signaling information based on a given link code.
     * 
     * @param linkCode - The link code to fetch signaling for.
     * @returns The signaling information for the user.
     */
    public async GetSignal(linkCode: string): Promise<Signal> {
        const { auth } = this;

        try {
            const res = await fetch(`${Signaling.SIGNAL_WITH_LINKCODE_URL}${linkCode}`, {
                method: "GET",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${auth.AccessToken}`,
                },
            });

            if (!res.ok) {
                throw new Error(`Failed to fetch signal: ${res.status}`);
            }

            const signal: Signal = await res.json();
            return signal;
        } catch (error) {
            console.error("Error while getting signal:", error);
            throw error;
        }
    }

    /**
     * Forwards a signal to the specified link code.
     * 
     * @param linkCode - The link code to forward the signal to.
     * @param signal - The signal to forward.
     */
    public async ForwardSignal(linkCode: string, signal: Signal): Promise<void> {
        const { auth } = this;

        const res = await fetch(`${Signaling.SIGNAL_WITH_LINKCODE_URL}${linkCode}`, {
            method: "POST",
            body: JSON.stringify(signal),
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${auth.AccessToken}`,
            },
        });

        if (!res.ok) {
            throw new Error(`Failed to forward signal: ${res.status}`);
        }
    }

    /**
     * Removes a signal based on the provided link code.
     * 
     * @param linkCode - The link code to remove.
     * @returns The HTTP status code of the deletion request.
     */
    public RemoveSignal(): void {
        const { setSignalController } = this;

        if (setSignalController && !setSignalController.signal.aborted) {
            try {
                setSignalController.abort();
            }
            catch { }
        }
    }
}
