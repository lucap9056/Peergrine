/**
 * 定義基礎消息類
 */
export class UserInfoMessage {
    type: "USER_INFO";
    data: {
        user_name: string;
    };

    /**
     * 創建一個新的 BaseMessage 實例
     * @param userName - 用戶名
     */
    constructor(user_name: string) {
        this.type = "USER_INFO"; // 設置消息類型
        this.data = { user_name }; // 設置用戶名
    }
}

/**
 * 定義用戶名變更消息類
 */
export class ChangeUserNameMessage {
    type: "CHANGE_USER_NAME";
    data: string;

    /**
     * 創建一個新的 ChangeUserNameMessage 實例
     * @param user_name - 新的用戶名
     */
    constructor(user_name: string) {
        this.type = "CHANGE_USER_NAME"; // 設置消息類型
        this.data = user_name; // 設置新的用戶名
    }
}

/**
 * 消息類型聯合
 */
type Message = UserInfoMessage | ChangeUserNameMessage;

export default Message;
