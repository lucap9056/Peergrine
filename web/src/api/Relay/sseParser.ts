import pako from "pako";

interface RawSessionData {
    client_id: string;
    public_key: string;
}

export interface SessionData {
    client_id: string;
    public_key: CryptoKey;
}

interface RawMessageData {
    sender_id: string;
    message: string;
}

export interface MessageData {
    sender_id: string;
    message: string;
}

export interface Message {
    event: string;
    data: string;
}

interface SSEParserData {
    key_name: string;
    key_format: Exclude<KeyFormat, "jwk">;
    hash_name: string;
    private_key: CryptoKey;
}

export default class SSEParser {
    public static ParseResponse(raw: Uint8Array): Message {
        const message: Message = {
            event: "",
            data: "",
        };

        const decoder = new TextDecoder();
        const data = decoder.decode(raw);

        // Split the raw SSE data by the two new lines separating each part
        for (const line of data.split(/\n\n/)) {
            if (line.startsWith('data: ')) {
                message.data = line.slice(6).replace(/\n*$/, '');  // Removing trailing newlines from the data
            } else if (line.startsWith('event: ')) {
                message.event = line.slice(7);  // Extract the event
            }
        }

        return message;
    }

    private readonly keyName: string;
    private readonly keyFormat: Exclude<KeyFormat, "jwk">;
    private readonly hashName: string;
    private readonly privateKey: CryptoKey;

    constructor(data: SSEParserData) {
        this.keyName = data.key_name;
        this.keyFormat = data.key_format;
        this.hashName = data.hash_name;
        this.privateKey = data.private_key;
    }

    // Encode the message data by encrypting it and then compressing it
    public async EncodeMessageData(key: CryptoKey, content: string): Promise<string> {
        const { keyName } = this;

        // Encrypt and compress the content
        const messageBytes = await crypto.subtle.encrypt(
            { name: keyName },
            key,
            pako.gzip(content)
        );

        // Convert encrypted bytes to base64 for transmission
        const messageBase64 = btoa(String.fromCharCode(...new Uint8Array(messageBytes)));

        return messageBase64;
    }

    // Decode the message data by decompressing and decrypting it
    public async DecodeMessageData(data: string): Promise<MessageData> {
        const { keyName, privateKey } = this;

        const { sender_id, message }: RawMessageData = JSON.parse(data);

        // Convert the base64 message to bytes
        const messageBytes = Uint8Array.from(atob(message), c => c.charCodeAt(0));

        // Decrypt the message bytes
        const decrypted = await crypto.subtle.decrypt({ name: keyName }, privateKey, messageBytes);

        // Decompress the decrypted bytes and return the original message
        const decompressed = pako.ungzip(new Uint8Array(decrypted), { to: "string" });

        return {
            sender_id,
            message: decompressed,
        };
    }

    // Encode session data by compressing the public key and converting it to base64
    public async EncodeSessionData(key: CryptoKey): Promise<string> {
        const { keyFormat } = this;

        // Export the public key
        const publicKeyBytes = await crypto.subtle.exportKey(keyFormat, key);

        // Compress and convert the public key bytes to base64
        return btoa(String.fromCharCode(...pako.gzip(publicKeyBytes)));
    }

    // Decode the session data by decompressing the public key and importing it
    public async DecodeSessionData(rawStr: string): Promise<SessionData> {
        const rawSessionData: RawSessionData = JSON.parse(rawStr);
        const { keyFormat, keyName, hashName } = this;
        const { client_id } = rawSessionData;

        // Convert the base64 public key to bytes
        const publicKeyBytes = Uint8Array.from(atob(rawSessionData.public_key), c => c.charCodeAt(0));

        // Import the public key after decompressing it
        const public_key = await crypto.subtle.importKey(
            keyFormat,
            pako.ungzip(publicKeyBytes),
            {
                name: keyName,
                hash: {
                    name: hashName,
                },
            },
            true,
            ["encrypt"]
        );

        return {
            client_id,
            public_key,
        };
    }
}
