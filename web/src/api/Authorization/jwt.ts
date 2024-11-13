/**
 * 定義授權資訊的接口
 */
interface Authorization {
    "refresh_token": string,
    "access_token": string,
    "expires_at": number
}

/**
 * 定義 JWT 的 Header 部分
 */
interface Header {
    alg: string; // 使用的加密算法，例如 "HS256"
    typ: string; // Token 類型，例如 "JWT"
}

/**
 * 定義 JWT 的 Payload 部分
 */
interface Payload {
    iss: string;
    exp: number;      // Token 的過期時間（Unix 時間戳）
    iat: number;      // Token 的發行時間（Unix 時間戳）
    user_id: string; // 客戶端 ID
}

/**
 * JWT Token 類別
 */
class JWT {
    private token: string;
    private header: Header;   // Token 的 Header 部分
    private payload: Payload; // Token 的 Payload 部分
    private signature: string; // Token 的 Signature 部分

    /**
     * 創建一個 Token 實例
     * 
     * @param token - JWT 字符串
     * @throws {Error} 當 token 格式不正確時拋出錯誤
     */
    constructor(token: string) {
        this.token = token;

        const parts = token.split('.');

        // 檢查 Token 是否包含三部分
        if (parts.length !== 3) {
            throw new Error('Invalid token format');
        }

        try {
            // 解析 Token 的 Header 和 Payload 部分
            this.header = this.parseBase64<Header>(parts[0]);
            this.payload = this.parseBase64<Payload>(parts[1]);
            this.signature = parts[2];
        } catch (error) {
            throw new Error('Invalid token content');
        }
    }

    public get Token() {
        return this.token;
    }

    /**
     * 獲取 token 的 Header 部分
     * 
     * @returns Token 的 Header 部分
     */
    public get Header(): Header {
        return this.header;
    }

    /**
     * 獲取 token 的 Payload 部分
     * 
     * @returns Token 的 Payload 部分
     */
    public get Payload(): Payload {
        return this.payload;
    }

    /**
     * 獲取 token 的 Signature 部分
     * 
     * @returns Token 的 Signature 部分
     */
    public get Signature(): string {
        return this.signature;
    }

    /**
     * 從 Base64 字符串中解析 JSON
     * 
     * @param base64 - Base64 編碼的字符串
     * @returns 解析後的對象
     * @throws {Error} 當 Base64 字符串無法解析為 JSON 時拋出錯誤
     */
    private parseBase64<T>(base64: string): T {
        try {
            const json = atob(base64); // 將 Base64 編碼的字符串解碼為 JSON 字符串
            return JSON.parse(json) as T; // 解析 JSON 字符串為對象
        } catch (error) {
            throw new Error('Failed to parse Base64 string');
        }
    }
}

export default JWT;

export type {
    Authorization,
    Payload
};