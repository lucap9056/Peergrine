<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from 'vue';

import SideOptionsComponent from "@Components/SideOptions/index.vue";
import HeaderComponent from "@Components/Header/index.vue";
import ChatComponent from "@Components/Chat/index.vue";

import NotificationsComponent from "@Components/Notifications/index.vue";
import FileViewerComponent from "@Components/FileViewer/index.vue";
import LoadingComponent from "@Components/Loading/index.vue";

import { loadingManager } from "@Components/Loading";
import FileManager from "@Src/storage/files";
import Client from "@Src/client";
import Chat from "@Components/Chat";

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

        onMounted(() => {

            const loading = loadingManager.Add();

            const c = new Client();

            c.Authorization.on("AuthorizationStateChanged", (e) => {
                if (!c.Profile.ClientName) {
                    const name = e.detail.user_id.substring(0, 8);
                    c.Profile.UpdateClientName(name);
                }
                loading.Remove();
                const f = new FileManager(c);
                const ch = new Chat(c, f);

                client.value = c;
                fileManager.value = f;
                chat.value = ch;
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
        };
    }
});
</script>

<template>
    <template v-if="client && fileManager && chat">

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