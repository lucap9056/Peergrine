import { FileInfo, FileChunk } from "./file";


interface base {
    id: string
    type: string
    content: any
    timestamp: number
}

interface TextMessage extends base {
    type: "TEXT"
    content: {
        text: string
    }
}

class TextMessage {
    public static New = function (text: string): TextMessage {

        return {
            id: crypto.randomUUID(),
            type: "TEXT",
            content: { text },
            timestamp: Date.now()
        }

    }
}

interface FileMessage extends base {
    type: "FILE"
    content: FileInfo
}

class FileMessage {
    public static New = function (fileInfo: FileInfo): FileMessage {

        return {
            id: crypto.randomUUID(),
            type: "FILE",
            content: fileInfo,
            timestamp: Date.now()
        }

    }
}

interface FileRequestMessage extends base {
    type: "FILE_REQUEST"
    content: {
        file_id: string
        index?: number
    }
}

class FileRequestMessage {
    public static New = function (file_id: string, index?: number): FileRequestMessage {
        return {
            id: crypto.randomUUID(),
            type: "FILE_REQUEST",
            content: { file_id, index },
            timestamp: Date.now()
        }
    }
}

interface FileChunkMessage extends base {
    type: "FILE_CHUNK"
    content: FileChunk
}

class FileChunkMessage {
    public static New = function (fileChunk: FileChunk): FileChunkMessage {

        return {
            id: crypto.randomUUID(),
            type: "FILE_CHUNK",
            content: fileChunk,
            timestamp: Date.now()
        }

    }
}

interface ReceivedMessage extends base {
    type: "RECEIVED"
    content: {
        message_id: string;
    }
}

class ReceivedMessage {
    public static New = function (message_id: string): ReceivedMessage {

        return {
            id: crypto.randomUUID(),
            type: "RECEIVED",
            content: { message_id },
            timestamp: Date.now()
        }

    }
}


type Message = TextMessage | FileMessage | FileChunkMessage | FileRequestMessage | ReceivedMessage;

interface ChatMessage extends base {
    channel_id: string
    sender_id: string
    type: "CHAT_MESSAGE"
    content: Message
}

class ChatMessage {
    public static New = function (channel_id: string, sender_id: string, content: Message, id?: string): ChatMessage {

        const chatMessage: ChatMessage = {
            id: crypto.randomUUID(),
            type: "CHAT_MESSAGE",
            channel_id,
            sender_id,
            content,
            timestamp: Date.now(),
        }

        if (id) {
            chatMessage.id = id;
        }

        return chatMessage;

    }
}


export {
    TextMessage,
    FileMessage,
    FileRequestMessage,
    FileChunkMessage,
    ReceivedMessage,
}

export type {
    ChatMessage,
    Message
}

export default ChatMessage;