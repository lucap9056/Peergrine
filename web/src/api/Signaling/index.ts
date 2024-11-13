import Authorization from "@API/Authorization";

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
    sdp: string;
    candidates: RTCIceCandidate[];
};

export default class Signaling {
    private auth: Authorization;

    // URL constants
    private static readonly API_BASE_URL = "/api/signal/";
    private static readonly SIGNAL_URL = `${Signaling.API_BASE_URL}`;
    private static readonly SIGNAL_WITH_LINKCODE_URL = `${Signaling.API_BASE_URL}`;

    constructor(authorization: Authorization) {
        this.auth = authorization;
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

        return new Promise((resolve, reject) => {
            fetch(Signaling.SIGNAL_URL, {
                method: "POST",
                body: JSON.stringify(signal),
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${auth.AccessToken}`,
                },
            }).then(async (res) => {
                if (!res.body) {
                    throw new Error("Response body is empty");
                }
                if (!res.ok) {
                    throw new Error(await res.text());
                }
                return res.body.getReader();
            }).then(async (reader) => {
                const decoder = new TextDecoder();
                while (true) {
                    const { done, value } = await reader.read();
                    if (done) break;

                    const resultStr = decoder.decode(value);
                    const result = JSON.parse(resultStr);

                    if (result.link_code) {
                        resolve(result);
                    }

                    if (result.sdp) {
                        Target(result);
                    }
                }
            }).catch(reject);
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

        try {

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
        } catch (error) {
            console.error("Error while forwarding signal:", error);
        }
    }

    /**
     * Removes a signal based on the provided link code.
     * 
     * @param linkCode - The link code to remove.
     * @returns The HTTP status code of the deletion request.
     */
    public async RemoveSignal(linkCode: string): Promise<number> {
        const { auth } = this;

        try {
            const res = await fetch(`${Signaling.SIGNAL_WITH_LINKCODE_URL}${linkCode}`, {
                method: "DELETE",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${auth.AccessToken}`,
                },
            });

            return res.status; // Return the status code directly
        } catch (error) {
            console.error("Error while removing signal:", error);
            throw error;
        }
    }
}
