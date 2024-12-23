<script lang="ts">
import { defineComponent, ref } from 'vue';

import notificationManager, { Notification } from "./";

export default defineComponent({
    setup: () => {
        const notification = ref<Notification>();

        const notifications = notificationManager.Notifications;
        if (notifications.length > 0) {
            notification.value = notifications[0];
        }

        notificationManager.on("NotificationsChanged", (e) => {
            if (e.detail.length > 0) {
                notification.value = e.detail[0];
            } else {
                notification.value = undefined;
            }
        });

        const HandleRemoveAlert = (id?: string) => {
            if (id === undefined) return;

            notificationManager.RemoveNotification(id);
        }

        return {
            notification,
            Notification,
            HandleRemoveAlert
        }
    }
});

</script>
<template>
    <template v-if="notification">
        <div class="notification_container">
            <div class="notification" :key="notification.Id" :data-type="notification.Type">
                <div class="notification_content">{{ notification.Text }}</div>
                <template v-if="notification.AllButtons.length > 0">
                    <button v-for="button of notification.AllButtons" @click="button.TriggerClick">
                        {{ button.Text }}
                    </button>
                </template>
                <div v-else-if="notification.Type === Notification.TYPE.ALERT" class="notification_remove"
                    @click="() => HandleRemoveAlert(notification?.Id)">
                    <ion-icon name="close"></ion-icon>
                </div>
            </div>
        </div>
    </template>
</template>

<style lang="scss" scoped>
.notification_container {
    position: fixed;
    right: 0;
    bottom: 0;
    display: flex;
    flex-flow: column-reverse;
    padding: 10px;
    pointer-events: none;

    .notification {
        position: relative;
        display: flex;
        flex-flow: row;
        background-color: #925a45;
        padding: 3px 10px;
        align-items: center;
        pointer-events: all;
        border-radius: 4px;
        box-shadow: 0 0 3px #5a3324;
        cursor: default;

        &[data-type="ERROR"] {
            background-color: #d14a4a;
        }

        .notification_content {
            line-height: 24px;
            font-size: 18px;
        }

        button {
            background-color: #ccc;
        }

        .notification_remove {
            opacity: 0;
            width: 24px;
            height: 24px;
            font-size: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
            transition-duration: 125ms;
        }

        &:hover .notification_remove {
            opacity: 1;
            scale: 1.1;
        }
    }
}
</style>