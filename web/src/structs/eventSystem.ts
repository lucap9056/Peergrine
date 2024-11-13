class BaseEventSystem<EventDefinitions> {
    private events: { [K in keyof EventDefinitions]?: Array<(payload: EventDefinitions[K]) => void> } = {};

    /**
     * Registers an event listener.
     * 
     * @param event - The name of the event to listen to.
     * @param listener - The callback function to handle the event.
     */
    public on<K extends keyof EventDefinitions>(event: K, listener: (payload: EventDefinitions[K]) => void) {
        if (!this.events[event]) {
            this.events[event] = [];
        }
        (this.events[event] as Array<(payload: EventDefinitions[K]) => void>).push(listener);
    }

    /**
     * Unregisters an event listener.
     * 
     * @param event - The name of the event to remove the listener from.
     * @param listener - The callback function to be removed.
     */
    public off<K extends keyof EventDefinitions>(event: K, listener: (payload: EventDefinitions[K]) => void) {
        if (!this.events[event]) return;

        this.events[event] = (this.events[event] as Array<(payload: EventDefinitions[K]) => void>).filter(l => l !== listener);
    }

    /**
     * Emits an event to notify all registered listeners.
     * 
     * @param event - The name of the event to emit.
     * @param value - The payload to send with the event.
     */
    protected emit<K extends keyof EventDefinitions>(event: K, value: EventDefinitions[K]) {
        if (!this.events[event]) return;

        for (const listener of this.events[event] as Array<(payload: EventDefinitions[K]) => void>) {
            listener(value);
        }
    }
}

export default BaseEventSystem;
