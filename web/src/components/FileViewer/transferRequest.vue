<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from "vue";
import FileTransfer, { FileTransferEvent, FileRequest } from "@Src/storage/files/fileTransfer";

export default defineComponent({
    props: {
        fileTransfer: {
            type: Object as () => FileTransfer,
            required: true
        },
    },
    setup: ({ fileTransfer }) => {
        const FileRequests = ref<{ [fileId: string]: FileRequest }>({});

        const TransferRequestedHandler = (e: FileTransferEvent<"TransferRequested">) => {
            const { id } = e.detail.info;
            FileRequests.value[id] = e.detail;
        };

        onMounted(() => {
            fileTransfer.on("TransferRequested", TransferRequestedHandler);
        });

        onUnmounted(() => {
            fileTransfer.off("TransferRequested", TransferRequestedHandler);
        });

        const FormatFileSize = (bytes: number) => {
            const UNITS = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            let unitIndex = 0;

            while (bytes >= 1024 && unitIndex < UNITS.length - 1) {
                bytes /= 1024;
                unitIndex++;
            }

            return `${bytes.toFixed(2)} ${UNITS[unitIndex]}`;
        };

        const HandleRemoveFileRequest = (fileId: string) => {
            delete FileRequests.value[fileId];
        };

        const HandleAllowFileRequest = (fileId: string) => {
            fileTransfer.AllowFileRequest(fileId);
            HandleRemoveFileRequest(fileId);
        };

        return {
            FileRequests,
            FormatFileSize,
            HandleRemoveFileRequest,
            HandleAllowFileRequest
        };
    }
});
</script>

<template>
    <div class="request_file_container">
        <div class="request_file" v-for="{ channel, info } in FileRequests" :key="channel.id">
            <div class="request_file_message">"{{ channel.target.name || channel.target.id }}" requested the file: {{
                info.name }} ({{
                    FormatFileSize(info.size) }})</div>
            <div class="request_file_options">
                <div class="request_file_option" @click="HandleRemoveFileRequest(info.id)">Deny</div>
                <div class="request_file_option" @click="() => HandleAllowFileRequest(info.id)">Allow</div>
            </div>
        </div>
    </div>
</template>

<style lang="scss" scoped>
.request_file_container {
    position: fixed;
    right: 0;
    bottom: 0;
    margin: 20px 10px;
    display: flex;

    .request_file {
        position: relative;
        background-color: #713e2e;
        display: flex;
        flex-flow: column;
        min-width: 240px;
        box-shadow: 0 2px 8px 3px #8b6041;
        border-radius: 5px;
        padding: 5px 15px;

        .request_file_options {
            display: flex;
            flex-flow: row;
            padding: 5px;
            gap: 5px;

            .request_file_option {
                background-color: #f6eacb;
                text-align: center;
                border-radius: 3px;
                flex: 1;
                color: #5b3f2d;
                font-weight: bold;
                cursor: default;
                user-select: none;
                transition-duration: 125ms;

                &:hover {
                    background-color: #d0c3a1;
                }

                &:active {
                    background-color: #c1b087;
                }
            }
        }
    }
}
</style>
