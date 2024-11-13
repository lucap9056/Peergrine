import { Message } from "@Src/storage/message";
import BaseEventSystem from "@Src/structs/eventSystem";

import Authorization from "@API/Authorization";
import Signaling from "@API/Signaling";

import RTCManager, { RTCConn } from "@Src/client/RTCManager";
import ClientProfile from "@Src/client/ClientProfile";
import RelayManager from "@Src/client/RelayManager";

type ManagerNames = "RTC" | "RELAY";

export interface User {
    id: string;
    name: string;
    online: boolean;
    manager: ManagerNames;
}

export class User {
    public static readonly MANAGER_NAMES = class {
        public static readonly RTC: ManagerNames = "RTC";
        public static readonly RELAY: ManagerNames = "RELAY";
    }

    constructor(manager: ManagerNames, id: string, name?: string, online: boolean = true) {
        if (!name) {
            name = id.substring(0, 8);
        }

        Object.assign(this, {
            id,
            name,
            online,
            manager
        });
    }
}

export interface Channel {
    id: string;
    target: User;
}

export class Channel {
    constructor(client: ClientProfile, target: User) {
        this.id = [client.ClientId, target.id].sort().join("-");
        this.target = target;
    }
}

type EventDefinitions = {
    "ChannelAppended": { detail: Channel };
    "ChannelRemoved": { detail: Channel };
    "ChannelStatusChanged": { detail: Channel };
    "FocusChannelChanged": { detail: Channel };
    "MessageAppended": { detail: { channel: Channel, message: Message } };
}

export type ClientEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

class Client extends BaseEventSystem<EventDefinitions> {
    public readonly Profile: ClientProfile;
    public readonly Relay: RelayManager;
    public readonly Authorization: Authorization;
    public readonly Rtc: RTCManager;
    private ChannelsMap: Map<string, Channel>;
    private FocusedChannelId = "";

    constructor() {
        super();
        const channels = new Map<string, Channel>();

        const authorization = new Authorization();
        const profile = new ClientProfile(authorization);
        const relay = new RelayManager(authorization);
        const signaling = new Signaling(authorization);
        const rtcManager = new RTCManager(profile, signaling);

        {
            const RTCUser = (conn: RTCConn): User => {
                const { targetId, targetName, online } = conn;
                const user = new User(User.MANAGER_NAMES.RTC, targetId, targetName, online);
                return user;
            }

            rtcManager.on("UserAppended", (e) => {
                const user = RTCUser(e.detail);
                const channel = new Channel(profile, user);

                e.detail.SetConnId(channel.id);
                channels.set(channel.id, channel);

                this.emit("ChannelAppended", { detail: channel });
                this.SetFocusChannel(channel.id);
            });

            rtcManager.on("UserStatusChanged", (e) => {
                const user = RTCUser(e.detail);
                const channel = new Channel(profile, user);

                if (channels.has(channel.id)) {
                    channels.set(channel.id, channel);
                    this.emit("ChannelStatusChanged", { detail: channel });
                }
            });

            rtcManager.on("MessageAppended", (e) => {
                const { user, message } = e.detail;
                const channel = channels.get(user.ConnId);

                if (!channel) {
                    throw new Error("Channel not found");
                }

                channels.delete(channel.id);
                channels.set(channel.id, channel);

                this.emit("MessageAppended", {
                    detail: { channel, message }
                });
            });
        }

        {
            const RelayUser = (userId: string): User => {
                const user = new User(User.MANAGER_NAMES.RELAY, userId);
                return user;
            }

            relay.on("UserAppended", (e) => {
                const { userId } = e.detail;
                const user = RelayUser(userId);
                const channel = new Channel(profile, user);

                channels.set(channel.id, channel);

                this.emit("ChannelAppended", { detail: channel });
                this.SetFocusChannel(channel.id);
            });

            relay.on("MessageAppended", (e) => {
                const { sender, message } = e.detail;
                let channel = channels.get(sender.channelId);

                if (!channel) {
                    const user = RelayUser(sender.userId);
                    channel = new Channel(profile, user);
                    channels.set(channel.id, channel);
                    this.emit("ChannelAppended", { detail: channel });
                } else {
                    channels.delete(channel.id);
                    channels.set(channel.id, channel);
                }

                this.emit("MessageAppended", {
                    detail: { channel, message }
                });
            });
        }

        this.ChannelsMap = channels;
        this.Authorization = authorization;
        this.Profile = profile;
        this.Relay = relay;
        this.Rtc = rtcManager;
    }

    public get FocusedChannel(): Channel | undefined {
        const { ChannelsMap, FocusedChannelId } = this;

        if (FocusedChannelId === "") {
            return undefined;
        }

        return ChannelsMap.get(FocusedChannelId);
    }

    public SetFocusChannel(channelId: string = ""): void {
        const channel = this.ChannelsMap.get(channelId);

        if (channel && channel.id !== this.FocusedChannelId) {
            this.FocusedChannelId = channelId;
            this.emit("FocusChannelChanged", { detail: channel });
        }
    }

    public get Channels(): Channel[] {
        return Array.from(this.ChannelsMap.values());
    }

    public RemoveChannel(id: string): void {
        const channel = this.ChannelsMap.get(id);

        if (channel) {
            this.ChannelsMap.delete(id);
            this.emit("ChannelRemoved", { detail: channel });
        }
    }

    public Close(): void {
        // Implement close functionality if necessary
    }
}

export default Client;
