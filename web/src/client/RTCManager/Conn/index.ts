import { Message } from "@Src/storage/message";
import ConnStateMessage, { UserInfoMessage } from "./stateMessage";
import BaseEventSystem from "@Src/structs/eventSystem";
import { Signal } from "@API/Signaling";

type ConnEventDefinitions = {
    "Ready": { detail: Conn };
    "Close": { detail: Conn };
    "ErrorOccurred": { detail: { conn: Conn, error: Error } };
    "SignalChanged": { detail: Signal };
    "UserStatusChanged": { detail: ConnStateMessage };
    "MessageAppended": { detail: { user: Conn, message: Message } };
}

export type ConnEvent<T extends keyof ConnEventDefinitions> = ConnEventDefinitions[T];

type DataType = {
    message: Message;
    state: ConnStateMessage;
};

type Channels = { [channelName: string]: RTCDataChannel };

type State = "INITIAL" | "OFFER_READY" | "ANSWER_READY" | "CONNECTING" | "CONNECTED" | "DISCONNECTED";

interface Conn {
    targetId: string;
    targetName: string;
    online: boolean;
}

class Conn extends BaseEventSystem<ConnEventDefinitions> {
    public static CHANNELS = class {
        public static MESSAGE: keyof DataType = "message";
        public static STATE: keyof DataType = "state";
    }

    private clientId: string;
    private clientName: string;
    private connId: string = "";
    private conn: RTCPeerConnection;
    private channels: Channels = {};

    public static STATUS = class {
        public static INITIAL: State = "INITIAL";
        public static OFFER_READY: State = "OFFER_READY";
        public static ANSWER_READY: State = "ANSWER_READY";
        public static CONNECTING: State = "CONNECTING";
        public static CONNECTED: State = "CONNECTED";
        public static DISCONNECTED: State = "DISCONNECTED";
    }

    private state: State = Conn.STATUS.INITIAL;

    constructor(clientId: string, clientName: string = "") {
        super();
        this.clientId = clientId;
        this.clientName = clientName;
        
        const conn = new RTCPeerConnection();

        conn.addEventListener("iceconnectionstatechange", () => {
            if (conn.iceConnectionState === "disconnected") {
                this.Close();
                this.state = Conn.STATUS.DISCONNECTED;
            }
        });

        conn.addEventListener("signalingstatechange", () => {
            
        });

        conn.addEventListener("datachannel", (e) => {
            this.Channel(e.channel);
        });

        const candidates: RTCIceCandidate[] = [];
        conn.addEventListener("icecandidate", (e) => {
            if (e.candidate) {
                candidates.push(e.candidate);
            } else if (conn.localDescription) {

                const { clientId, state } = this;
                const sdp = conn.localDescription.sdp;

                if (state !== Conn.STATUS.ANSWER_READY && state !== Conn.STATUS.OFFER_READY) {
                    return;
                }

                this.state = Conn.STATUS.CONNECTING;

                setTimeout(() => {

                    const signal: Signal = { client_id: clientId, sdp, candidates };
                    this.emit("SignalChanged", { detail: signal });

                }, 1000);
            }
        });

        this.conn = conn;
    }

    public Send<Channel extends keyof DataType>(channelName: Channel, data: DataType[Channel]): void {
        const channel = this.channels[channelName];
        if (channel) {
            const dataStr = JSON.stringify(data);
            if (channel.readyState === "open") {
                channel.send(dataStr);
            } else {
                channel.onopen = () => {
                    channel.send(dataStr);
                }
            }
        }
    }

    private Channel(channel: RTCDataChannel) {
        const { label } = channel;
        this.channels[label] = channel;

        channel.addEventListener("open", () => {
            switch (label) {
                case Conn.CHANNELS.STATE: {
                    const base = new UserInfoMessage(this.clientName);
                    this.Send(Conn.CHANNELS.STATE, base);
                    break;
                }
            }
        });

        switch (label) {
            case Conn.CHANNELS.MESSAGE:
                channel.addEventListener("message", (e) => this.ConnMessage(e));
                break;
            case Conn.CHANNELS.STATE:
                channel.addEventListener("message", (e) => this.ConnStatus(e));
                break;
        }
    }

    private ConnMessage(e: MessageEvent<string>) {
        const message: Message = JSON.parse(e.data);
        this.emit("MessageAppended", { detail: { user: this, message } });
    }

    private ConnStatus(e: MessageEvent<string>) {
        const status: ConnStateMessage = JSON.parse(e.data);

        switch (status.type) {
            case "USER_INFO":
                this.targetName = status.data.user_name;
                this.state = Conn.STATUS.CONNECTED;
                this.emit("Ready", { detail: this });
                break;
            case "CHANGE_USER_NAME":
                this.targetName = status.data;
                break;
        }

        this.emit("UserStatusChanged", { detail: status });
    }

    public async CreateOffer() {
        const { conn } = this;

        const defaultChannels = [Conn.CHANNELS.MESSAGE, Conn.CHANNELS.STATE];
        defaultChannels.forEach(channelName => {
            const channel = conn.createDataChannel(channelName);
            this.Channel(channel);
        });

        const offer = await conn.createOffer();
        await conn.setLocalDescription(offer);

        this.state = Conn.STATUS.OFFER_READY;
    }

    public async SetConnectionTarget({ client_id, sdp, candidates }: Signal) {
        this.targetId = client_id;

        const description = new RTCSessionDescription({ type: "answer", sdp: decodeURIComponent(sdp) });
        await this.conn.setRemoteDescription(description);

        for (const icecandidate of candidates) {
            await this.conn.addIceCandidate(icecandidate);
        }

        this.state = Conn.STATUS.CONNECTING;
    }

    public async CreateAnswer({ client_id, sdp, candidates }: Signal) {
        const { conn } = this;

        this.targetId = client_id;

        try {
            const description = new RTCSessionDescription({ type: "offer", sdp: decodeURIComponent(sdp) });
            await conn.setRemoteDescription(description);

            for (const candidate of candidates) {
                await conn.addIceCandidate(candidate);
            }

            const answer = await conn.createAnswer();
            await conn.setLocalDescription(answer);
        } catch (e: any) {
            if (!(e instanceof Error)) {
                e = new Error(e);
            }
            this.emit("ErrorOccurred", { detail: { conn: this, error: e } });
            return;
        }

        this.state = Conn.STATUS.ANSWER_READY;
    }

    public SetConnId(id: string): void {
        this.connId = id;
    }

    public get ConnId(): string {
        return this.connId;
    }

    public Close() {
        this.conn.close();
        this.state = Conn.STATUS.DISCONNECTED;
        this.emit("Close", { detail: this });
    }
}

export default Conn;
