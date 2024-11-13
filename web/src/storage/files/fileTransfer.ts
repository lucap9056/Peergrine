import Client, { Channel, User } from "@Src/client";
import { CompleteFileWrapper } from "./structs";
import BaseEventSystem from "@Src/structs/eventSystem";
import { FileChunk } from "@Src/storage/message/file";

export interface FileRequest extends CompleteFileWrapper {
    channel: Channel;
    index?: number;
}

type EventDefinitions = {
    "TransferRequested": {
        detail: FileRequest;
    },
    "TransferProgressed": {
        detail: {
            fileId: string;
            owner: boolean;
            current: number;
            maximum: number;
        };
    },
    "TransferCompleted": {
        detail: { fileId: string };
    },
    "TransferDenied": {
        detail: { fileId: string };
    }
}

export type FileTransferEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

class FileTransfer extends BaseEventSystem<EventDefinitions> {
    public static readonly CHUNK_SIZE: number = 16 * 1024;

    private receivedFileRequests: { [fileId: string]: FileRequest } = {};
    private cancelRequests: { [fileId: string]: boolean } = {};

    private client: Client;

    constructor(client: Client) {
        super();
        this.client = client;
    }

    public SendFileRequest(fileId: string, index?: number): void {
        const { client } = this;
        const channel = client.FocusedChannel;

        if (!channel) {
            throw new Error("No channel found.");
        }

        switch (channel.target.manager) {
            case User.MANAGER_NAMES.RTC: {
                client.Rtc.SendFileRequestMessage(channel.target.id, fileId, index);
                break;
            }
        }
    }

    public HandleFileRequest(channel: Channel, fileW: CompleteFileWrapper, index?: number): void {
        const request = { ...CompleteFileWrapper.HasFile(fileW), channel, index };

        this.receivedFileRequests[fileW.info.id] = request;

        this.emit("TransferRequested", {
            detail: request
        });
    }

    public AllowFileRequest(fileId: string): void {
        const request = this.receivedFileRequests[fileId];

        if (request) {
            delete this.receivedFileRequests[fileId];
            this.SendFile(request);
        }
    }

    public DenyFileRequest(fileId: string): void {
        const { client, receivedFileRequests } = this;
        const request = receivedFileRequests[fileId];

        if (request) {
            delete receivedFileRequests[fileId];

            const { channel, info } = request;

            switch (channel.target.manager) {
                case User.MANAGER_NAMES.RTC: {
                    const targetId = channel.target.id;

                    const chunk = new FileChunk(info.id, -1, 0, "");
                    client.Rtc.SendFileChunkMessage(targetId, chunk);
                    break;
                }
            }
        }
    }

    public CancelFileRequest(fileId: string): void {
        this.cancelRequests[fileId] = true;
    }

    public HandleFileChunk(chunk: FileChunk): void {
        const { file_id, index, total } = chunk;

        if (index === -1) {
            this.emit("TransferDenied", { detail: { fileId: file_id } });
            return;
        }
        
        this.emit("TransferProgressed", {
            detail: {
                fileId: file_id,
                owner: false,
                current: index,
                maximum: total
            }
        });

        if (index === total - 1) {
            this.emit("TransferCompleted", { detail: { fileId: file_id } });
        }
    }

    private SendFile(request: FileRequest): void {
        const { client, cancelRequests } = this;
        const { channel, file, info } = request;
        const fileId = info.id;

        delete cancelRequests[fileId];

        switch (channel.target.manager) {
            case User.MANAGER_NAMES.RTC: {
                const total = Math.ceil(file.size / FileTransfer.CHUNK_SIZE);
                let index = request.index || 0;

                const ReadSlice = () => {
                    if (cancelRequests[fileId]) {
                        const chunk = new FileChunk(info.id, -1, total, "");
                        client.Rtc.SendFileChunkMessage(channel.target.id, chunk);
                        setTimeout(() => {
                            this.emit("TransferDenied", { detail: { fileId } });
                        }, 100);
                        return;
                    }

                    const offset = index * FileTransfer.CHUNK_SIZE;
                    const end = offset + FileTransfer.CHUNK_SIZE;
                    const slice = file.slice(offset, end);

                    slice.arrayBuffer().then((buffer) => {
                        const byteArray = new Uint8Array(buffer);

                        let binaryStr = '';
                        byteArray.forEach((byte) => {
                            binaryStr += String.fromCharCode(byte);
                        });

                        const chunk: FileChunk = new FileChunk(info.id, index, total, binaryStr);

                        client.Rtc.SendFileChunkMessage(channel.target.id, chunk);

                        index++;
                        if (index < total) {
                            ReadSlice();

                            this.emit("TransferProgressed", {
                                detail: {
                                    fileId,
                                    owner: true,
                                    current: index,
                                    maximum: total
                                }
                            });
                        } else {
                            this.emit("TransferCompleted", {
                                detail: { fileId }
                            });
                        }
                    });
                }

                ReadSlice();
            }
        }
    }
}

export default FileTransfer;
