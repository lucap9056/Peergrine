import { FileInfo, FileChunk } from "@Src/storage/message/file";
import { ChatMessage } from "@Src/storage/message";

type State = "INITIAL" | "READY" | "CLEARED";

interface StorageMessage extends ChatMessage {
    clientId: string
}

interface StorageFileChunk extends FileChunk {
    clientId: string
}

class DBChatStorage {
    public static ForceUpdateOnNext = () => {
        document.cookie = "indexedDBRebuild=true; path=/";
    }

    public static readonly NAME = "IndexedDB";
    public readonly NAME = DBChatStorage.NAME;

    public static DB_NAME = "Peergrine";
    public static STORE_NAMES = class {
        public static readonly MESSAGE = "PeergrineChatStore";
        public static readonly FILE_CHUNK = "PeergrineFileChunkStore";
    }

    public static STATUS = class {
        public static readonly INITIAL: State = "INITIAL";
        public static readonly READY: State = "READY";
        public static readonly CLEARED: State = "CLEARED";
    }

    public static INDEXES = class {
        public static readonly MESSAGE_TIMESTAMP = "MessageTimestampIndex";
        public static readonly MESSAGE_CHANNEL = "MessageChannelIndex";
        public static readonly FILE_CLIENT = "FileClientIndex";
        public static readonly FILE_INDEX = "FileIndex";
    }

    public static STORE_LOSE_ERROR = new Error("storeLose");

    private db?: IDBDatabase;
    private state: State = DBChatStorage.STATUS.INITIAL;

    private clientId: string;

    public static New = function (clientId: string): Promise<DBChatStorage> {
        return new Promise((resolve, reject) => {
            const storage = new DBChatStorage(clientId);

            const forceUpdate = document.cookie.includes("indexedDBRebuild=true");
            document.cookie = "indexedDBRebuild=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";

            storage.InitDatabase(forceUpdate).then(() => {
                resolve(storage);
            }).catch(reject);
        });
    }

    constructor(clientId: string) {
        this.clientId = clientId;
    }

    private InitDatabase(forceUpdate: boolean = false): Promise<void> {
        return new Promise((resolve, reject) => {
            if (forceUpdate) {
                console.log("Force update database");
            }
            const version = forceUpdate ? new Date().getTime() : undefined;
            const request = indexedDB.open(DBChatStorage.DB_NAME, version);

            request.onupgradeneeded = (e) => {
                const req = e.target as IDBRequest;
                try {
                    this.RebuildDatabase(req.result);
                } catch (err) {
                    if (req.transaction) {
                        req.transaction.abort();
                    }
                    reject(err);
                }
            };

            request.onsuccess = () => {
                const db = request.result;
                const stores = [
                    DBChatStorage.STORE_NAMES.MESSAGE,
                    DBChatStorage.STORE_NAMES.FILE_CHUNK
                ];

                for (const storeName of stores) {
                    if (!db.objectStoreNames.contains(storeName)) {
                        reject(DBChatStorage.STORE_LOSE_ERROR);
                        return;
                    }
                }

                this.db = db;
                this.state = DBChatStorage.STATUS.READY;
                resolve();
            };

            request.onerror = () => reject(request.error);
        });
    }

    public RebuildDatabase(db: IDBDatabase): void {
        console.log("Rebuilding database");
        const stores = [
            DBChatStorage.STORE_NAMES.MESSAGE,
            DBChatStorage.STORE_NAMES.FILE_CHUNK
        ];

        stores.forEach(storeName => {
            if (db.objectStoreNames.contains(storeName)) {
                db.deleteObjectStore(storeName);
            }
        });

        const chatStore = db.createObjectStore(DBChatStorage.STORE_NAMES.MESSAGE, { keyPath: "id" });
        chatStore.createIndex(DBChatStorage.INDEXES.MESSAGE_TIMESTAMP, "timestamp", { unique: false });
        chatStore.createIndex(DBChatStorage.INDEXES.MESSAGE_CHANNEL, ["clientId", "channel_id"], { unique: false });

        const fileStore = db.createObjectStore(DBChatStorage.STORE_NAMES.FILE_CHUNK, { keyPath: "id" });
        fileStore.createIndex(DBChatStorage.INDEXES.FILE_CLIENT, ["clientId"], { unique: false });
        fileStore.createIndex(DBChatStorage.INDEXES.FILE_INDEX, ["file_id", "index"], { unique: false });
    }

    public async AddMessage(chatMessage: ChatMessage): Promise<void> {
        const { db, clientId } = this;

        return new Promise((resolve, reject) => {
            if (!db) {
                reject(new Error("Database not initialized"));
                return;
            }

            const transaction = db.transaction([
                DBChatStorage.STORE_NAMES.MESSAGE
            ], "readwrite");

            switch (chatMessage.content.type) {
                case "TEXT":
                case "FILE": {
                    const message: StorageMessage = { ...chatMessage, clientId };
                    const store = transaction.objectStore(DBChatStorage.STORE_NAMES.MESSAGE);
                    const req = store.put(message);
                    req.onsuccess = () => resolve();
                    req.onerror = () => reject(req.error);
                    break;
                }
            }
        });
    }

    public GetMessagesForChannel(channelId: string): Promise<ChatMessage[]> {
        const { clientId, db } = this;

        return new Promise((resolve, reject) => {

            if (!db) {
                reject(new Error("Database not initialized"));
                return;
            }

            const transaction = db.transaction([DBChatStorage.STORE_NAMES.MESSAGE], "readonly");
            const store = transaction.objectStore(DBChatStorage.STORE_NAMES.MESSAGE);
            const index = store.index(DBChatStorage.INDEXES.MESSAGE_CHANNEL);
            const range = IDBKeyRange.only([clientId, channelId]);

            const messages: ChatMessage[] = [];

            const cursorRequest = index.openCursor(range);
            cursorRequest.onsuccess = (e) => {
                const cursor = (e.target as IDBRequest<IDBCursorWithValue>).result;
                if (cursor) {
                    messages.push(cursor.value);
                    cursor.continue();
                } else {
                    messages.sort((a, b) => a.timestamp - b.timestamp);
                    resolve(messages);
                }
            };
        });
    }

    public AddFileChunk(chunk: FileChunk): Promise<void> {
        const { db, clientId } = this;

        const data: StorageFileChunk = { ...chunk, clientId };

        return new Promise((resolve, reject) => {
            if (!db) throw new Error("Database not initialized");

            const transaction = db.transaction([
                DBChatStorage.STORE_NAMES.FILE_CHUNK
            ], "readwrite");

            const store = transaction.objectStore(DBChatStorage.STORE_NAMES.FILE_CHUNK);
            const req = store.add(data);
            req.onsuccess = () => resolve();
            req.onerror = () => reject(req.error);
        });
    }

    public GetFileChunkCount(fileId: string): Promise<number> {
        const { db } = this;
        return new Promise((resolve, reject) => {

            if (!db) {
                reject(new Error("Database not initialized"));
                return;
            }

            const transaction = db.transaction(DBChatStorage.STORE_NAMES.FILE_CHUNK, "readonly");
            const store = transaction.objectStore(DBChatStorage.STORE_NAMES.FILE_CHUNK);
            const index = store.index(DBChatStorage.INDEXES.FILE_INDEX);
            const range = IDBKeyRange.bound([fileId, 0], [fileId, Number.MAX_SAFE_INTEGER]);

            const req = index.count(range);
            req.onsuccess = () => {
                resolve(req.result);
            };

            req.onerror = () => {
                reject(req.error);
            };

        });
    }

    public DownloadFileFromStorage(fileInfo: FileInfo): Promise<void> {
        const { db } = this;
        const { id, name } = fileInfo;

        return new Promise((resolve, reject) => {

            if (!db) {
                reject(new Error("Database not initialized"));
                return;
            }

            const stream = new ReadableStream({
                start(controller) {
                    const transaction = db.transaction(DBChatStorage.STORE_NAMES.FILE_CHUNK, "readonly");
                    const store = transaction.objectStore(DBChatStorage.STORE_NAMES.FILE_CHUNK);
                    const index = store.index(DBChatStorage.INDEXES.FILE_INDEX);
                    const range = IDBKeyRange.bound([id, 0], [id, Number.MAX_SAFE_INTEGER]);
                    const cursorRequest = index.openCursor(range);


                    cursorRequest.onsuccess = () => {
                        const cursor = cursorRequest.result;

                        if (cursor) {

                            const { data } = cursor.value as FileChunk;
                            const bytes = new Uint8Array(data.length);
                            for (let i = 0; i < data.length; i++) {
                                bytes[i] = data.charCodeAt(i);
                            }

                            controller.enqueue(bytes);
                            cursor.continue();
                        } else {
                            controller.close();
                        }
                    }

                    cursorRequest.onerror = (e) => {
                        controller.close();
                        reject(e);
                    };

                }
            });

            const blob = new Response(stream).blob();

            blob.then((fileBlob) => {
                const url = URL.createObjectURL(fileBlob);

                const a = document.createElement('a');
                a.href = url;
                a.download = name;
                a.style.display = "none";
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);

                URL.revokeObjectURL(url);
                resolve();
            });

        });
    }

    public ClearFile(fileId: string): Promise<void> {
        const { db } = this;
        return new Promise((resolve, reject) => {

            if (!db) {
                reject(new Error("Database not initialized"));
                return;
            }

            const transaction = db.transaction(DBChatStorage.STORE_NAMES.FILE_CHUNK, "readwrite");
            const store = transaction.objectStore(DBChatStorage.STORE_NAMES.FILE_CHUNK);
            const index = store.index(DBChatStorage.INDEXES.FILE_INDEX);
            const range = IDBKeyRange.bound([fileId, 0], [fileId, Number.MAX_SAFE_INTEGER]);
            const cursorRequest = index.openCursor(range);
            cursorRequest.onsuccess = () => {
                const cursor = cursorRequest.result;
                if (cursor) {
                    store.delete(cursor.primaryKey);
                    cursor.continue();
                } else {
                    resolve();
                }
            }

            cursorRequest.onerror = () => reject(cursorRequest.error);
        });
    }

    public async ClearClientData(): Promise<void> {
        const { db, clientId } = this;

        if (!db) {
            throw new Error("Database not initialized");
        }

        const transaction = db.transaction(
            [
                DBChatStorage.STORE_NAMES.MESSAGE,
                DBChatStorage.STORE_NAMES.FILE_CHUNK
            ], "readwrite");

        const clearMessage: Promise<void> = new Promise((resolve, reject) => {
            const messageStore = transaction.objectStore(DBChatStorage.STORE_NAMES.MESSAGE);
            const messageIndex = messageStore.index(DBChatStorage.INDEXES.MESSAGE_CHANNEL);
            const messageRange = IDBKeyRange.only([clientId]);
            const messageCursorRequest = messageIndex.openCursor(messageRange);

            messageCursorRequest.onsuccess = () => {
                const cursor = messageCursorRequest.result;

                if (cursor) {
                    messageStore.delete(cursor.primaryKey);
                } else {
                    resolve();
                }
            };

            messageCursorRequest.onerror = () => reject(messageCursorRequest.error);
        });

        const clearFile: Promise<void> = new Promise((resolve, reject) => {
            const fileStore = transaction.objectStore(DBChatStorage.STORE_NAMES.FILE_CHUNK);
            const fileIndex = fileStore.index(DBChatStorage.INDEXES.FILE_CLIENT);
            const fileRange = IDBKeyRange.only([clientId]);
            const fileCursorRequest = fileIndex.openCursor(fileRange);

            fileCursorRequest.onsuccess = () => {
                const cursor = fileCursorRequest.result;

                if (cursor) {
                    fileStore.delete(cursor.primaryKey);
                } else {
                    resolve();
                }
            }

            fileCursorRequest.onerror = () => reject(fileCursorRequest.error);
        });

        await Promise.all([clearMessage, clearFile]);
    }

    public ClearAllData(): Promise<void> {
        const { db } = this;

        if (!db) {
            throw new Error("Database not initialized");
        }

        db.close();

        return new Promise((resolve, reject) => {

            const req = indexedDB.deleteDatabase(DBChatStorage.DB_NAME);
            req.onsuccess = () => {
                resolve();
            };
            req.onerror = () => reject(req.error);

        });
    }

    public get State(): State {
        return this.state;
    }

}

export type {
    FileInfo,
    FileChunk,
}

export default DBChatStorage;
