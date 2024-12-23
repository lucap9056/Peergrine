<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from 'vue';

import SideOptionsComponent from "@Components/SideOptions/index.vue";
import HeaderComponent from "@Components/Header/index.vue";
import ChatComponent from "@Components/Chat/index.vue";

import NotificationsComponent from "@Components/Notifications/index.vue";
import FileViewerComponent from "@Components/FileViewer/index.vue";
import LoadingComponent from "@Components/Loading/index.vue";

import { loadingManager, Loading } from "@Components/Loading";
import FileManager from "@Src/storage/files";
import Client from "@Src/client";
import Chat from "@Components/Chat";
import notificationManager, { Notification } from '@Components/Notifications';

export default defineComponent({
    components: {
        SideOptionsComponent,
        HeaderComponent,
        ChatComponent,
        NotificationsComponent,
        FileViewerComponent,
        LoadingComponent,
    },
    setup() {
        const client = ref<Client>();
        const fileManager = ref<FileManager>();
        const chat = ref<Chat>();
        const connected = ref<boolean>(false);

        onMounted(() => {

            let loading: Loading | undefined = loadingManager.Add();

            let notification: Notification | undefined = new Notification(Notification.TYPE.ALERT, "Connection is being initialized...");
            notificationManager.AddNotification(notification);

            const c = new Client();

            c.Authorization.on("AuthorizationStateChanged", (e) => {
                if (!c.Profile.ClientName) {
                    const name = e.detail.user_id.substring(0, 8);
                    c.Profile.UpdateClientName(name);
                }

                if (loading) {
                    loading.Remove();
                    loading = undefined;
                }

                if (notification) {
                    notification.Remove();
                    notification = undefined;
                }

                const f = new FileManager(c);
                const ch = new Chat(c, f);

                client.value = c;
                fileManager.value = f;
                chat.value = ch;
                connected.value = true;
            });

            c.Authorization.on("ConnectionClosed", () => {
                connected.value = false;

                if (loading === undefined) {
                    loading = loadingManager.Add();
                }

                if (notification) {
                    notification.Remove();
                }

                notification = new Notification(Notification.TYPE.ALERT, "Connection has been disconnected.");
                notificationManager.AddNotification(notification);
            });

        });

        onUnmounted(() => {
            if (client.value) {
                client.value.Close();
            }
        });

        return {
            client,
            fileManager,
            chat,
            connected,
        };
    }
});
</script>

<template>
    <template v-if="client && fileManager && chat && connected">

        <HeaderComponent :profile="client.Profile"></HeaderComponent>

        <div class="main">
            <SideOptionsComponent :client></SideOptionsComponent>
            <router-view :client :chat :fileManager></router-view>
            <ChatComponent :client :chat :fileManager></ChatComponent>
        </div>

        <FileViewerComponent :fileManager></FileViewerComponent>

    </template>

    <NotificationsComponent></NotificationsComponent>
    <LoadingComponent></LoadingComponent>
</template>

<style lang="scss" scoped>
.main {
    display: flex;
    flex: 1;
}
</style>