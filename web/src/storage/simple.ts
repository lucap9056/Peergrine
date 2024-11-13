import { ChatMessage } from "@Src/storage/message";
import { FileChunk, FileInfo, StorageFile } from "./message/file";

class SimpleChatStorage {
    public static readonly NAME = "Memory";
    public readonly NAME = SimpleChatStorage.NAME;

    private messages: { [messageId: string]: ChatMessage } = {};
    private channelMessages: { [channelId: string]: string[] } = {};
    private files: { [fileId: string]: StorageFile } = {};

    public async AddMessage(chatMessage: ChatMessage) {
        const { messages, channelMessages, files } = this;
        const chatMessages = channelMessages[chatMessage.channel_id];
        const messageId = chatMessage.id;
        const message = chatMessage.content;

        switch (message.type) {
            case "FILE": {
                const { content } = message;
                files[content.id] = new StorageFile(content);

                messages[messageId] = chatMessage;

                if (chatMessages) {
                    chatMessages.push(messageId);
                } else {
                    channelMessages[chatMessage.channel_id] = [messageId];
                }
                break;
            }
            case "TEXT": {
                messages[messageId] = chatMessage;

                if (chatMessages) {
                    chatMessages.push(messageId);
                } else {
                    channelMessages[chatMessage.channel_id] = [messageId];
                }
                break;
            }
        }
    }

    public async GetMessagesForChannel(connId: string): Promise<ChatMessage[]> {
        const { channelMessages, messages } = this;
        return (channelMessages[connId] || []).map(id => messages[id]);
    }

    public async AddFileChunk(chunk: FileChunk): Promise<void> {
        const { files } = this;
        const { file_id } = chunk;
        if (files[file_id]) {
            files[file_id].chunks.push(chunk);
        }
    }

    public async GetFileChunkCount(id: string): Promise<number> {
        const file = this.files[id];
        if (file && file.chunks) {
            return file.chunks.length;
        }
        return 0;
    }

    public DownloadFileFromStorage(info: FileInfo): Promise<void> {
        const { id, name } = info;
        const file = this.files[id];

        return new Promise((resolve, reject) => {

            if (!file) {
                reject(new Error('File not found'));
                return;
            }

            const { chunks } = file;

            const stream = new ReadableStream({
                start(controller) {
                    chunks.sort((a, b) => a.index - b.index).forEach(({ data }) => {
                        const bytes = new Uint8Array(data.length);

                        for (let i = 0; i < data.length; i++) {
                            bytes[i] = data.charCodeAt(i);
                        }

                        controller.enqueue(bytes);
                    });
                    controller.close();
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

    public async ClearFile(fileId: string): Promise<void> {
        const { files } = this;

        const file = files[fileId];
        if (file) {
            file.chunks = [];
        }
    }
}

export default SimpleChatStorage;
