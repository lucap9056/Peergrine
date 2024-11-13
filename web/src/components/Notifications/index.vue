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

        return {
            notification
        }
    }
});

</script>
<template>
    <template v-if="notification">
        <div class="notification_container">
            <div class="notification" :key="notification.Id">
                <div class="notification_content">{{ notification.Text }}</div>
                <template v-if="notification.AllButtons.length > 0">
                    <button v-for="button of notification.AllButtons" @click="button.TriggerClick">
                        {{ button.Text }}
                    </button>
                </template>
            </div>
        </div>
    </template>
</template>
