import BaseEventSystem from "@Src/structs/eventSystem";

type ButtonEventDefinitions = {
    "Clicked": { detail: Notification }
};
export type NotificationButtonManagerEvent<T extends keyof ButtonEventDefinitions> = ButtonEventDefinitions[T];

class NotificationButton extends BaseEventSystem<ButtonEventDefinitions> {
    private readonly buttonText: string;
    private autoRemoveNotification: boolean = true;
    private associatedNotification?: Notification;

    constructor(buttonText: string) {
        super();
        this.buttonText = buttonText;
    }

    public get Text(): string {
        return this.buttonText;
    }

    public SetAutoRemove(autoRemove: boolean = true): NotificationButton {
        this.autoRemoveNotification = autoRemove;
        return this;
    }

    public SetNotification(notification: Notification): void {
        this.associatedNotification = notification;
    }

    public TriggerClick(): void {
        const { associatedNotification, autoRemoveNotification } = this;
        if (!associatedNotification) return;

        this.emit("Clicked", { detail: associatedNotification });

        if (autoRemoveNotification) {
            associatedNotification.Remove();
        }
    }
}

type NotificationType = "NORMAL" | "ALERT" | "ERROR";

interface Notification {
    text: string;
}

class Notification {
    public static TYPE = class {
        public static readonly NORMAL: NotificationType = "NORMAL";
        public static readonly ALERT: NotificationType = "ALERT";
        public static readonly ERROR: NotificationType = "ERROR";
    }

    private readonly notificationType: NotificationType;
    private readonly notificationText: string;
    private buttons: NotificationButton[] = [];
    private notificationId: string = "";
    private manager?: NotificationManager;

    constructor(notificationType: NotificationType, notificationText: string, buttons: NotificationButton[] = []) {
        this.notificationType = notificationType;
        this.notificationText = notificationText;
        this.buttons = buttons;

        buttons.forEach(button => {
            button.SetNotification(this);
        });
    }

    public get Type(): NotificationType {
        return this.notificationType;
    }

    public get Text(): string {
        return this.notificationText;
    }

    public set Id(notificationId: string) {
        this.notificationId = notificationId;
    }

    public get Id(): string {
        return this.notificationId;
    }

    public AddButton(button: NotificationButton): void {
        button.SetNotification(this);
        this.buttons.push(button);
    }

    public get HasButtons(): boolean {
        return this.buttons.length > 0;
    }

    public get AllButtons(): NotificationButton[] {
        return this.buttons;
    }

    public set Manager(manager: NotificationManager) {
        this.manager = manager;
    }

    public Remove(): void {
        if (!this.manager) return;
        this.manager.RemoveNotification(this.notificationId);
    }
}

type ManagerEventDefinitions = {
    "NotificationsChanged": { detail: Notification[] }
};
export type NotificationManagerEvent<T extends keyof ManagerEventDefinitions> = ManagerEventDefinitions[T];

class NotificationManager extends BaseEventSystem<ManagerEventDefinitions> {
    private notificationsList: Notification[] = [];

    public AddNotification(notification: Notification): void {
        notification.Id = crypto.randomUUID();
        notification.Manager = this;

        let autoRemoveDelay = 0;
        switch (notification.Type) {
            case Notification.TYPE.NORMAL: {
                autoRemoveDelay = 3000;
                break;
            }
            case Notification.TYPE.ERROR: {
                autoRemoveDelay = 5000;
                break;
            }
        }

        this.notificationsList.push(notification);

        if (!notification.HasButtons && autoRemoveDelay !== 0) {
            setTimeout(() => this.RemoveNotification(notification.Id), autoRemoveDelay);
        }

        this.emit("NotificationsChanged", { detail: this.notificationsList });
    }

    public RemoveNotification(NotificationId: string): void {
        const Index = this.notificationsList.findIndex((Notification) => Notification.Id === NotificationId);

        if (Index > -1) {
            this.notificationsList = [
                ...this.notificationsList.slice(0, Index),
                ...this.notificationsList.slice(Index + 1)
            ];

            this.emit("NotificationsChanged", { detail: this.notificationsList });
        }
    }

    public get Notifications(): Notification[] {
        return this.notificationsList;
    }
}

export {
    NotificationButton,
    Notification,
    NotificationManager,
};
export type {
    NotificationType,
};

const notificationManager = new NotificationManager();

export default notificationManager;
