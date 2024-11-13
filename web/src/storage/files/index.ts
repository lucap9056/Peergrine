import BaseEventSystem from "@Src/structs/eventSystem";
import { FileChunk, FileInfo } from "@Src/storage/message/file";
import Client from "@Src/client";
import FileTransfer from "./fileTransfer";
import { FileWrapper, CompleteFileWrapper } from "./structs";

export interface FileStorage {
    NAME: string
    AddFileChunk: (chunk: FileChunk) => Promise<void>;
    GetFileChunkCount: (fileId: string) => Promise<number>;
    DownloadFileFromStorage: (info: FileInfo) => Promise<void>;
    ClearFile: (fileId: string) => Promise<void>;
}

type FileStorages = {
    [StorageType in FileStorage['NAME']]: FileStorage
}

export type FileExistenceState = "NotFound" | "PartiallyExists" | "Exists";

type EventDefinitions = {
    "FileViewed": { detail: FileInfo };
    "FileStorageChanged": { detail: string }
    "FileStorageAppended": { detail: string }
    "FileStorageRemoved": { detail: string }
}

export type FileManagerEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

class FileManager extends BaseEventSystem<EventDefinitions> {
    public static FileExistenceState = class {
        public static readonly NotFound: FileExistenceState = "NotFound";
        public static readonly PartiallyExists: FileExistenceState = "PartiallyExists";
        public static readonly Exists: FileExistenceState = "Exists";
    }

    public static CHUNK_SIZE: number = 16 * 1024;

    private files: { [fileId: string]: FileWrapper } = {};
    public storages: FileStorages = {};
    private storage?: FileStorage;

    public readonly fileTransfer: FileTransfer;

    constructor(client: Client) {
        super();
        this.fileTransfer = new FileTransfer(client);
    }

    public AddFile(file: File): FileInfo {
        const info = new FileInfo(file);

        this.files[info.id] = { file, info };

        return info;
    }

    public AddFileInfo(info: FileInfo): void {
        this.files[info.id] = { info };
    }

    public async AddFileChunk(chunk: FileChunk): Promise<void> {
        if (this.storage) {
            await this.storage.AddFileChunk(chunk)
        }
        this.fileTransfer.HandleFileChunk(chunk);
    }

    public GetFileById(id: string): FileWrapper | undefined {
        return this.files[id];
    }

    public async GetFileChunkCountInStorage(id: string): Promise<number> {
        const { storage, files } = this;

        const file = files[id];

        if (file && storage) {
            return await storage.GetFileChunkCount(id);
        }

        return 0;
    }

    public async CheckFileExistenceInStorage(id: string): Promise<FileExistenceState> {
        const { storage, files } = this;

        const file = files[id];

        if (file && storage) {
            const count = await storage.GetFileChunkCount(id);

            if (count === 0) {
                return FileManager.FileExistenceState.NotFound;
            } else if (file.info.size <= count * FileManager.CHUNK_SIZE) {
                return FileManager.FileExistenceState.Exists;
            } else {
                return FileManager.FileExistenceState.PartiallyExists;
            }
        }

        return FileManager.FileExistenceState.NotFound;
    }

    public async DownloadFile(id: string): Promise<void> {
        const { storage, files } = this;

        const file = files[id];
        if (storage && file) {
            return await storage.DownloadFileFromStorage(file.info);
        }
    }

    public async ClearFileFromStorage(id: string): Promise<void> {
        const { storage } = this;

        if (storage) {
            storage.ClearFile(id);
        }
    }

    public DisplayFile(id: string): void {
        const file = this.files[id];

        if (file) {
            this.emit("FileViewed", { detail: file.info });
        }
    }

    public GetFileStorage<K extends keyof FileStorages>(name: K): FileStorage | undefined {
        return this.storages[name];
    }

    public get FileStorageNames(): string[] {
        return Object.keys(this.storages);
    }

    public get FileStorageName(): string {
        return this.storage?.NAME || "";
    }

    public SetFileStorage<K extends keyof FileStorages>(name: K): void {
        if (!this.storages[name]) {
            throw new Error(`FileStorage with the name "${name}" does not exist in storages. Available storage names: ${Object.keys(this.storages).join(', ')}.`);
        }
        this.storage = this.storages[name];
        this.emit("FileStorageChanged", { detail: name });
    }

    public AddFileStorage(storage: FileStorage): void {
        this.storages[storage.NAME] = storage;
        if (!this.storage) {
            this.SetFileStorage(storage.NAME);
        }
        this.emit("FileStorageAppended", { detail: storage.NAME });
    }

    public RemoveFileStorage<K extends keyof FileStorages>(name: K): void {
        if (!this.storages[name]) {
            throw new Error(`FileStorage with the name "${name}" does not exist in storages. Available storage names: ${Object.keys(this.storages).join(', ')}.`);
        }

        delete this.storages[name];

        if (this.storage && this.storage.NAME === name) {

            const storages = Object.values(this.storages);

            if (storages.length > 0) {
                this.storage = storages[0];
                this.emit("FileStorageChanged", { detail: storages[0].NAME });
            } else {
                this.storage = undefined;
                this.emit("FileStorageChanged", { detail: "" });
            }

        }

        this.emit("FileStorageRemoved", { detail: name });
    }

    public get IsFileStorageEnabled(): boolean {
        return this.storage !== undefined;
    }
}

export default FileManager;
export {
    CompleteFileWrapper
};
export type {
    FileWrapper
};
