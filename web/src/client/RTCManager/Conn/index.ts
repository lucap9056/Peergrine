import { Message } from "@Src/storage/message";
import ConnStateMessage, { UserInfoMessage } from "./stateMessage";
import BaseEventSystem from "@Src/structs/eventSystem";
import { Signal } from "@API/Signaling";



type DataType = {
    message: Message;
    state: ConnStateMessage;
};

type Channels = { [channelName: string]: RTCDataChannel };

type State = "INITIAL" | "OFFER_READY" | "ANSWER_READY" | "CONNECTING" | "CONNECTED" | "DISCONNECTED" | "FAILED";

interface Conn {
    targetId: string;
    targetName: string;
    online: boolean;
}

type ConnEventDefinitions = {
    "Ready": { detail: Conn };
    "Close": { detail: Conn };
    "ErrorOccurred": { detail: { conn: Conn, error: Error } };
    "SignalChanged": { detail: Signal };
    "UserStatusChanged": { detail: ConnStateMessage };
    "ConnStateChanged": { detail: { state: State } };
    "MessageAppended": { detail: { user: Conn, message: Message } };
}

export type ConnEvent<T extends keyof ConnEventDefinitions> = ConnEventDefinitions[T];
class Conn extends BaseEventSystem<ConnEventDefinitions> {
    public static readonly CHANNELS = class {
        public static readonly MESSAGE: keyof DataType = "message";
        public static readonly STATE: keyof DataType = "state";
    }

    private clientId: string;
    private clientName: string;
    private connId: string = "";
    private conn: RTCPeerConnection;
    private channels: Channels = {};

    public static readonly STATUS = class {
        public static readonly INITIAL: State = "INITIAL";
        public static readonly OFFER_READY: State = "OFFER_READY";
        public static readonly ANSWER_READY: State = "ANSWER_READY";
        public static readonly CONNECTING: State = "CONNECTING";
        public static readonly CONNECTED: State = "CONNECTED";
        public static readonly DISCONNECTED: State = "DISCONNECTED";
        public static readonly FAILED: State = "FAILED";
    }

    private state: State = Conn.STATUS.INITIAL;
    private signal?: Signal;

    constructor(config: RTCConfiguration, client_id: string, client_name: string = "") {
        super();
        this.clientId = client_id;
        this.clientName = client_name;
        this.online = false;

        const conn = new RTCPeerConnection(config);

        const candidates: RTCIceCandidate[] = [];

        conn.addEventListener("iceconnectionstatechange", () => {

            switch (conn.iceConnectionState) {
                case "disconnected": {
                    this.Close();
                    break;
                }
                case "failed": {
                    this.SetState(Conn.STATUS.FAILED);
                    break;
                }
            }
        });

        conn.addEventListener("signalingstatechange", async () => {

            switch (conn.signalingState) {
                case "have-local-offer": {

                    if (conn.localDescription) {

                        const { clientId } = this;

                        const sdp = conn.localDescription.sdp;
                        this.signal = { client_id: clientId, channel_id: "", sdp, candidates };
                        this.SetState(Conn.STATUS.OFFER_READY);
                    }
                    break;
                }
                case "have-remote-offer": {
                    const answer = await conn.createAnswer();
                    await conn.setLocalDescription(answer);

                    if (conn.localDescription) {
                        const sdp = conn.localDescription.sdp;
                        this.signal = { client_id, channel_id: "", sdp, candidates };
                        this.SetState(Conn.STATUS.ANSWER_READY);
                    }

                    break;
                }
            }
        });

        conn.addEventListener("datachannel", (e) => {
            this.Channel(e.channel);
        });

        conn.addEventListener("icecandidate", (e) => {
            if (e.candidate) {
                candidates.push(e.candidate);
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
                this.SetState(Conn.STATUS.CONNECTED);
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

    }

    public async SetConnectionTarget({ client_id, sdp, candidates }: Signal) {
        this.targetId = client_id;

        const description = new RTCSessionDescription({ type: "answer", sdp });
        await this.conn.setRemoteDescription(description);

        for (const icecandidate of candidates) {
            await this.conn.addIceCandidate(icecandidate);
        }

        this.SetState(Conn.STATUS.CONNECTING);
    }

    public async CreateAnswer({ client_id, sdp, candidates }: Signal) {
        const { conn } = this;

        this.targetId = client_id;

        const description = new RTCSessionDescription({ type: "offer", sdp });
        await conn.setRemoteDescription(description);

        for (const candidate of candidates) {
            await conn.addIceCandidate(candidate);
        }
    }

    public get Signal(): Signal | undefined {
        return this.signal;
    }

    private SetState(state: State): void {
        this.online = state === Conn.STATUS.CONNECTED;
        this.state = state;

        this.emit("ConnStateChanged", { detail: { state } });
    }

    public get State(): State {
        return this.state;
    }

    public SetConnId(id: string): void {
        this.connId = id;
    }

    public get ConnId(): string {
        return this.connId;
    }

    public Close() {
        this.conn.close();
        this.SetState(Conn.STATUS.DISCONNECTED);
        this.emit("Close", { detail: this });
    }
}

export default Conn;
