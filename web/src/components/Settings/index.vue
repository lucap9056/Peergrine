<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from "vue"

import Chat, { ChatEvent } from "@Components/Chat";
import FileManager, { FileManagerEvent } from "@Src/storage/files";
import Select from "@Components/Settings/select.vue";
import DBChatStorage from "@Src/storage/indexedDB";

export default defineComponent({
    components: {
        Select
    },
    props: {
        chat: {
            type: Object as () => Chat,
            required: true
        },
        fileManager: {
            type: Object as () => FileManager,
            required: true
        }
    },
    setup: ({ chat, fileManager }) => {
        const chatStorageNames = ref<string[]>(chat.ChatStorageNames);
        const fileStorageNames = ref<string[]>(fileManager.FileStorageNames);

        const chatStorage = ref<string>(chat.ChatStorageName);
        const fileStorage = ref<string>(fileManager.FileStorageName);

        const ChatStorageChangedHandler = (e: ChatEvent<"ChatStorageChanged">) => {
            chatStorage.value = e.detail;
        }

        const ChatStorageAppendedHandler = (e: ChatEvent<"ChatStorageAppended">) => {
            chatStorageNames.value.push(e.detail);
        }

        const ChatStorageRemovedHandler = (e: ChatEvent<"ChatStorageRemoved">) => {
            chatStorageNames.value = chat.ChatStorageNames;
            if (chatStorage.value === e.detail) {
                chatStorage.value = chat.ChatStorageName;
            }
        }

        const FileStorageChangedHandler = (e: FileManagerEvent<"FileStorageChanged">) => {
            fileStorage.value = e.detail;
        }

        const FileStorageAppendedHandler = (e: FileManagerEvent<"FileStorageAppended">) => {
            fileStorageNames.value.push(e.detail);
        }

        const FileStorageRemovedHandler = (e: FileManagerEvent<"FileStorageRemoved">) => {
            fileStorageNames.value = fileManager.FileStorageNames;
            if (fileStorage.value === e.detail) {
                fileStorage.value = fileManager.FileStorageName;
            }
        }

        onMounted(() => {
            chat.on("ChatStorageChanged", ChatStorageChangedHandler);
            chat.on("ChatStorageAppended", ChatStorageAppendedHandler);
            chat.on("ChatStorageRemoved", ChatStorageRemovedHandler);
            fileManager.on("FileStorageChanged", FileStorageChangedHandler);
            fileManager.on("FileStorageAppended", FileStorageAppendedHandler);
            fileManager.on("FileStorageRemoved", FileStorageRemovedHandler);
        });

        onUnmounted(() => {
            chat.off("ChatStorageChanged", ChatStorageChangedHandler);
            chat.off("ChatStorageAppended", ChatStorageAppendedHandler);
            chat.off("ChatStorageRemoved", ChatStorageRemovedHandler);
            fileManager.off("FileStorageChanged", FileStorageChangedHandler);
            fileManager.off("FileStorageAppended", FileStorageAppendedHandler);
            fileManager.off("FileStorageRemoved", FileStorageRemovedHandler);
        });

        const HandleChangeChatStorage = (v: string) => {
            chat.SetChatStorage(v);
        }

        const HandleChangeFileStorage = (v: string) => {
            fileManager.SetFileStorage(v);
        }

        const HandleIndexedDBClearClientData = () => {
            const storage = chat.GetChatStorage(DBChatStorage.NAME);
            if (storage) {
                (storage as DBChatStorage).ClearClientData().catch(console.error);
            }
        }

        const HandleIndexedDBClearAllData = () => {
            const storage = chat.GetChatStorage(DBChatStorage.NAME);
            if (storage) {
                (storage as DBChatStorage).ClearAllData().then(() => {
                    chat.RemoveChatStorage(DBChatStorage.NAME);
                    fileManager.RemoveFileStorage(DBChatStorage.NAME);
                }).catch(console.error);
            }
            DBChatStorage.ForceUpdateOnNext();
        }

        return {
            chatStorageNames,
            fileStorageNames,
            chatStorage,
            fileStorage,
            HandleChangeChatStorage,
            HandleChangeFileStorage,
            HandleIndexedDBClearClientData,
            HandleIndexedDBClearAllData,
        }
    }
})
</script>

<template>
    <div class="settings">

        <div class="settings_option">
            <div class="settings_label">Message Storeage</div>
            <Select :options="chatStorageNames" :key="chatStorage" :value="chatStorage"
                @update="HandleChangeChatStorage"></Select>
        </div>

        <div class="settings_option">
            <div class="settings_label">File Storeage</div>
            <Select :options="fileStorageNames" :key="fileStorage" :value="fileStorage"
                @update="HandleChangeFileStorage"></Select>
        </div>

        <div class="settings_option">
            <div class="settings_label">IndexedDB</div>
            <button class="settings_button" @click="HandleIndexedDBClearClientData">Clear Client Data</button>
            <button class="settings_button" @click="HandleIndexedDBClearAllData">Clear All Data</button>
        </div>

    </div>
</template>

<style lang="scss" scoped>
.settings {
    width: 208px;
    background-color: #654F3A;

    .settings_button {
        position: relative;
        margin: 5px;
        width: 80%;
        background-color: #e6ac83;
        color: #56321e;
        border: none;
        transition: 150ms;

        &:hover {
            transform: scale(1.02);
            background-color: #cb9670;
        }

        &:active {
            transform: scale(0.95);
            background-color: #bb8762;
        }
    }
}
</style>