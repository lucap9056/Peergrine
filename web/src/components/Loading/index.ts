import BaseEventSystem from "@Src/structs/eventSystem";

class Loading {
    private id: string;
    private manager: LoadingManager;

    constructor(id: string, manager: LoadingManager) {
        this.id = id;
        this.manager = manager;
    }

    public Remove(): void {
        this.manager.RemoveLoadingById(this.id);
    }
}

type EventDefinitions = {
    "StateChanged": { detail: boolean }
};

export type LoadingManagerEvent<T extends keyof EventDefinitions> = EventDefinitions[T];

class LoadingManager extends BaseEventSystem<EventDefinitions> {
    private loadings: { [loadingId: string]: Loading } = {};
    private state: boolean = false;

    public Add(): Loading {
        const id = crypto.randomUUID();
        const loading = new Loading(id, this);

        this.loadings[id] = loading;

        this.UpdateLoadingState();
        return loading;
    }

    public RemoveLoadingById(loadingId: string): void {
        delete this.loadings[loadingId];

        this.UpdateLoadingState();
    }

    private UpdateLoadingState(): void {
        const newState = Object.keys(this.loadings).length > 0;
        if (newState !== this.state) {
            this.state = newState;
            this.emit("StateChanged", { detail: newState });
        }
    }

    public get IsLoading(): boolean {
        return this.state;
    }
}

const loadingManager = new LoadingManager();

export default LoadingManager;
export {
    Loading,
    loadingManager
};
