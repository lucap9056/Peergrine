import BaseEventSystem from "@Src/structs/eventSystem";
import Authorization from "@API/Authorization";

// 定義事件名稱與其資料結構
type EventDefinitions = {
    "ClientNameChanged": { detail: string }
};

export type ClientProfileEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

export default class ClientProfile extends BaseEventSystem<EventDefinitions> {
    private auth: Authorization;
    private name: string = "";

    // 建構函數，傳入 Authorization 物件
    constructor(authorization: Authorization) {
        super();
        this.auth = authorization;
    }

    // 客戶端 ID，從授權資訊中取得
    public get ClientId(): string {
        return this.auth.Payload?.user_id || "";
    }

    // 客戶端名稱
    public get ClientName(): string {
        return this.name;
    }

    // 更新客戶端名稱並觸發事件
    public UpdateClientName(name: string): void {
        this.name = name;
        this.emit("ClientNameChanged", { detail: name });
    }
}
