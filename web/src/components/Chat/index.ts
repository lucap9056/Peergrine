import Client, { ClientEvent, User, Channel } from "@Src/client";
import BaseEventSystem from "@Src/structs/eventSystem";
import ChatMessage from "@Src/storage/message";

import SimpleChatStore from "@Src/storage/simple";
import DBChatStore from "@Src/storage/indexedDB";
import FileManager, { CompleteFileWrapper } from "@Src/storage/files";

interface ChatStorage {
    NAME: string
    AddMessage: (message: ChatMessage) => Promise<void>;
    GetMessagesForChannel: (connId: string) => Promise<ChatMessage[]>;
}

type ChatStorages = {
    [StorageType in ChatStorage['NAME']]: ChatStorage
}

type EventDefinitions = {
    "ChatStorageChanged": { detail: string }
    "ChatStorageAppended": { detail: string }
    "ChatStorageRemoved": { detail: string }
    "ChannelChanged": { detail: Channel };
    "MessagesUpdated": { detail: ChatMessage[] };
    "MessageAppended": { detail: ChatMessage };
    "ErrorOccurred": { detail: { message: string, error: Error } };
}

export type ChatEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

export default class Chat extends BaseEventSystem<EventDefinitions> {
    private client?: Client;
    private storages: ChatStorages = {};
    private storage: ChatStorage;
    private fileManager?: FileManager;
    private focusedChannel?: Channel;

    constructor(client: Client, fileManager: FileManager) {
        super();

        const simpleStorage = new SimpleChatStore();
        this.storages[SimpleChatStore.NAME] = simpleStorage;
        this.storage = simpleStorage;

        fileManager.AddFileStorage(simpleStorage);

        this.client = client;
        this.fileManager = fileManager;

        client.on("FocusChannelChanged", (e) => this.HandleFocusChannelChanged(e));
        client.on("ChannelStatusChanged", (e) => this.HandleChannelStatusChanged(e));
        client.on("MessageAppended", (e) => this.HandleMessageAppended(e));

        
        DBChatStore.New(client.Profile.ClientId).then((dbChatStorage) => {
            this.AddChatStorage(dbChatStorage);
            fileManager.AddFileStorage(dbChatStorage);
        }).catch((err) => {
            this.HandleError("", err);
        });

    }

    private HandleFocusChannelChanged(e: ClientEvent<"FocusChannelChanged">) {
        this.focusedChannel = e.detail;
        this.emit("ChannelChanged", { detail: e.detail });

        this.storage.GetMessagesForChannel(e.detail.id).then((messages) => {
            this.emit("MessagesUpdated", { detail: messages });
        });
    }

    private HandleChannelStatusChanged(e: ClientEvent<"ChannelStatusChanged">) {
        if (this.focusedChannel && this.focusedChannel.id === e.detail.id) {
            this.emit("ChannelChanged", { detail: e.detail });
        }
    }

    private HandleMessageAppended(e: ClientEvent<"MessageAppended">) {
        if (!this.client) {
            throw new Error("Client is not initialized");
        }

        const { client, storage, focusedChannel, fileManager } = this;
        const { channel, message } = e.detail;

        switch (message.type) {
            case "TEXT":
            case "FILE": {
                const { target } = channel;

                switch (target.manager) {
                    case User.MANAGER_NAMES.RTC:
                        client.Rtc.MessageReceived(target.id, message.id);
                }

                const chatMessage = ChatMessage.New(channel.id, target.id, message);
                storage.AddMessage(chatMessage).catch(console.error);

                if (focusedChannel && focusedChannel.id === channel.id) {
                    this.emit("MessageAppended", { detail: chatMessage });
                }

                if (message.type === "FILE" && fileManager) {
                    fileManager.AddFileInfo(message.content);
                }
                break;
            }
            case "FILE_REQUEST": {
                if (fileManager) {
                    const { file_id, index } = message.content;
                    const fileWrapper = fileManager.GetFileById(file_id);
                    if (fileWrapper && fileWrapper.file) {
                        fileManager.fileTransfer.HandleFileRequest(channel, CompleteFileWrapper.HasFile(fileWrapper), index);
                    }
                }
                break;
            }
            case "FILE_CHUNK": {
                if (fileManager) {
                    fileManager.AddFileChunk(message.content);
                    fileManager.fileTransfer.HandleFileChunk(message.content);
                }
                break;
            }
        }
    }

    public SendTextMessage(content: string): void {
        const { client, focusedChannel, storage } = this;

        if (client && focusedChannel) {
            const { ClientId } = client.Profile;
            const channelId = focusedChannel.id;
            const { id, manager } = focusedChannel.target;

            switch (manager) {
                case User.MANAGER_NAMES.RTC: {
                    const message = client.Rtc.SendTextMessage(id, content);
                    const chatMessage = ChatMessage.New(channelId, ClientId, message, message.id);
                    storage.AddMessage(chatMessage);
                    this.emit("MessageAppended", { detail: chatMessage });
                    break;
                }
                case User.MANAGER_NAMES.RELAY: {
                    client.Relay.SendMessage(id, content).then((message) => {
                        const chatMessage = ChatMessage.New(channelId, ClientId, message, message.id);
                        storage.AddMessage(chatMessage);
                        this.emit("MessageAppended", { detail: chatMessage });
                    });
                }
            }
        }
    }

    public SendFileMessage(file: File): void {
        const { client, focusedChannel, fileManager } = this;

        if (client && focusedChannel && fileManager) {
            const { ClientId } = client.Profile;
            const channelId = focusedChannel.id;
            const { id, manager } = focusedChannel.target;

            const fileInfo = fileManager.AddFile(file);

            switch (manager) {
                case User.MANAGER_NAMES.RTC: {
                    const message = client.Rtc.SendFileMessage(id, fileInfo);
                    const chatMessage = ChatMessage.New(channelId, ClientId, message, message.id);
                    this.emit("MessageAppended", { detail: chatMessage });
                }
            }
        }
    }

    public GetChatStorage<K extends keyof ChatStorages>(name: K): ChatStorage | undefined {
        return this.storages[name];
    }

    public get ChatStorageNames(): string[] {
        return Object.keys(this.storages);
    }

    public get ChatStorageName(): string {
        return this.storage.NAME;
    }

    public SetChatStorage<K extends keyof ChatStorages>(name: K): void {
        if (!this.storages[name]) {
            throw new Error(`ChatStorage with the name "${name}" does not exist in storages. Available storage names: ${Object.keys(this.storages).join(', ')}.`);
        }
        this.storage = this.storages[name];
        this.emit("ChatStorageChanged", { detail: name });
    }

    private AddChatStorage(storage: ChatStorage): void {
        this.storages[storage.NAME] = storage;
        this.emit("ChatStorageAppended", { detail: storage.NAME });
    }

    public RemoveChatStorage<K extends keyof ChatStorages>(name: K): void {
        if (!this.storages[name]) {
            throw new Error(`ChatStorage with the name "${name}" does not exist in storages. Available storage names: ${Object.keys(this.storages).join(', ')}.`);
        }

        if (name === SimpleChatStore.NAME) {
            throw new Error();
        }

        delete this.storages[name];

        if (this.storage.NAME === name) {

            const storages = Object.values(this.storages);

            if (storages.length > 0) {
                this.storage = storages[0];
            }

        }

        this.emit("ChatStorageRemoved", { detail: name });
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
}
