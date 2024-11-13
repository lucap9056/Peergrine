<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from "vue";
import FileManager, { FileManagerEvent, FileExistenceState } from "@Src/storage/files";
import { FileInfo } from "@Src/storage/message/file";
import FileTransferRequest from "./transferRequest.vue";
import { FileTransferEvent } from "@Src/storage/files/fileTransfer";
import { loadingManager } from "../Loading";

interface FileTransferProgressInfo {
    id: string;
    name: string;
    size: number;
    rate: string;
}

export default defineComponent({
    components: {
        FileTransferRequest
    },
    props: {
        fileManager: {
            type: Object as () => FileManager,
            required: true
        }
    },
    setup: ({ fileManager }) => {

        const fileStorageEnabled = ref<boolean>(fileManager.IsFileStorageEnabled);

        const fileInfo = ref<FileInfo>();
        const fileExistenceState = ref<FileExistenceState>(FileManager.FileExistenceState.NotFound);
        const fileOwner = ref<boolean>(false);
        const fileTransferProgressInfo = ref<FileTransferProgressInfo>();

        const FileStorageChangedHandler = (e: FileManagerEvent<"FileStorageChanged">) => {
            fileStorageEnabled.value = e.detail !== "";
        };

        const FileViewedHandler = (e: FileManagerEvent<"FileViewed">) => {
            fileInfo.value = e.detail;
            fileManager.CheckFileExistenceInStorage(e.detail.id).then((state) => {
                fileExistenceState.value = state;
            }).catch(console.error);
        };

        const TransferProgressedHandler = (e: FileTransferEvent<"TransferProgressed">) => {
            const { fileId, owner, current, maximum } = e.detail;
            const file = fileManager.GetFileById(fileId);

            if (file) {
                const { name, size } = file.info;

                const rate = (current / maximum * 100).toFixed() + "%";

                fileOwner.value = owner;
                fileTransferProgressInfo.value = { id: fileId, name, size, rate };
            }
        };

        const TransferCompletedHandler = (e: FileTransferEvent<"TransferCompleted">) => {
            const { fileId } = e.detail;

            if (fileTransferProgressInfo.value && fileTransferProgressInfo.value.id === fileId) {
                fileTransferProgressInfo.value = undefined;
            }
            if (fileInfo.value && fileInfo.value.id === fileId) {
                fileManager.CheckFileExistenceInStorage(fileId).then((state) => {
                    fileExistenceState.value = state;
                }).catch(console.error);
            }
        };

        onMounted(() => {
            fileManager.on("FileStorageChanged", FileStorageChangedHandler);
            fileManager.on("FileViewed", FileViewedHandler);

            fileManager.fileTransfer.on("TransferProgressed", TransferProgressedHandler);
            fileManager.fileTransfer.on("TransferCompleted", TransferCompletedHandler);
            fileManager.fileTransfer.on("TransferDenied", TransferCompletedHandler);
        });

        onUnmounted(() => {
            fileManager.off("FileStorageChanged", FileStorageChangedHandler);
            fileManager.off("FileViewed", FileViewedHandler);

            fileManager.fileTransfer.off("TransferProgressed", TransferProgressedHandler);
            fileManager.fileTransfer.off("TransferCompleted", TransferCompletedHandler);
            fileManager.fileTransfer.off("TransferDenied", TransferCompletedHandler);
        });

        const FormatFileSize = (bytes: number) => {
            const units = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            let unitIndex = 0;

            while (bytes >= 1024 && unitIndex < units.length - 1) {
                bytes /= 1024;
                unitIndex++;
            }

            return `${bytes.toFixed(2)} ${units[unitIndex]}`;
        };

        const HasFile = (): boolean => {
            if (fileInfo.value) {
                const fileId = fileInfo.value.id;
                const fileWrapper = fileManager.GetFileById(fileId);
                if (fileWrapper) {
                    return fileWrapper.file !== undefined;
                }
            }
            return false;
        };

        const HandleDownloadFile = async () => {
            if (fileInfo.value) {
                const loading = loadingManager.Add();
                try {
                    await fileManager.DownloadFile(fileInfo.value.id);
                }
                catch (err) {
                    console.error(err);
                }
                loading.Remove();
            }
        };

        const HandleClearFile = async () => {
            if (fileInfo.value) {
                const loading = loadingManager.Add();
                const { id } = fileInfo.value;
                try {
                    await fileManager.ClearFileFromStorage(id);
                } catch (err) {
                    console.error(err);
                }

                fileManager.CheckFileExistenceInStorage(id).then((state) => {
                    fileExistenceState.value = state;
                });
                loading.Remove();
            }
        };

        const HandleRequestFile = () => {
            if (fileInfo.value) {
                fileManager.fileTransfer.SendFileRequest(fileInfo.value.id);
            }
        };

        const HandleRetryFileRequest = () => {
            if (fileInfo.value) {
                const { id } = fileInfo.value;
                fileManager.GetFileChunkCountInStorage(id).then((count) => {
                    fileManager.fileTransfer.SendFileRequest(id, count);
                });
            }
        };

        const HandleCancelTransfer = () => {
            if (fileTransferProgressInfo.value) {
                fileManager.fileTransfer.CancelFileRequest(fileTransferProgressInfo.value.id);
            }
        };

        const HandleCloseFileViewer = () => {
            fileInfo.value = undefined;
        };

        return {
            fileInfo,
            fileExistenceState,
            fileOwner,
            fileTransferProgressInfo,
            fileManager,
            fileStorageEnabled,
            Status: FileManager.FileExistenceState,
            FormatFileSize,
            HasFile,
            HandleDownloadFile,
            HandleClearFile,
            HandleRequestFile,
            HandleRetryFileRequest,
            HandleCancelTransfer,
            HandleCloseFileViewer
        };
    }
});
</script>


<template>
    <div class="file_viewer_container" v-if="fileInfo !== undefined">
        <div class="file_viewer">
            <div class="file_viewer_file_name">{{ fileInfo.name }}</div>
            <div class="file_viewer_file_icon">
                <ion-icon name="document"></ion-icon>
            </div>
            <div class="file_viewer_file_size">{{ FormatFileSize(fileInfo.size) }}</div>

            <div v-if="!fileStorageEnabled" class="file_viewer_message">File Storage Not Enabled</div>

            <div class="file_viewer_options">
                <template v-if="fileStorageEnabled">
                    <div v-if="fileExistenceState === Status.NotFound" class="file_viewer_option"
                        @click="HandleRequestFile">
                        Request
                    </div>
                    <template v-else>
                        <div class="file_viewer_option" @click="HandleClearFile">Clear</div>
                        <div v-if="fileExistenceState === Status.Exists" class="file_viewer_option"
                            @click="HandleDownloadFile">
                            Download</div>
                        <div v-if="fileExistenceState === Status.PartiallyExists" class="file_viewer_option"
                            @click="HandleRetryFileRequest">
                            Retry Request</div>
                    </template>
                </template>
                <div class="file_viewer_option" @click="HandleCloseFileViewer">Close</div>
            </div>
        </div>
    </div>

    <FileTransferRequest :fileTransfer="fileManager.fileTransfer"></FileTransferRequest>

    <div class="file_transferring_container" v-if="fileTransferProgressInfo !== undefined">
        <div class="file_transferring">
            <div class="file_transferring_info">{{ fileTransferProgressInfo.name }}({{
                FormatFileSize(fileTransferProgressInfo.size)
                }})</div>
            <div class="file_transferring_progress_bar" :data-rate="fileTransferProgressInfo.rate">
                <div class="file_transferring_progress_rate" :style="{ width: fileTransferProgressInfo.rate }"></div>
            </div>
            <div v-if="fileOwner" class="file_transferring_cancel" @click="HandleCancelTransfer">Cancel</div>
        </div>
    </div>
</template>

<style lang="scss" scoped>
$progress_bar_clr1: #9191d0;
$progress_bar_clr2: #b799d5;

.file_viewer_container {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    justify-content: center;
    align-items: center;

    .file_viewer {
        background-color: #64493f;
        padding: 0.5rem;
        border-radius: 7px;
        display: flex;
        flex-flow: column;

        .file_viewer_file_icon {
            font-size: 4rem;
            line-height: 5rem;
        }

        .file_viewer_options {
            position: relative;
            display: flex;
            flex-flow: row;
            gap: 6px;
            padding: 0 20px;
            min-width: 180px;

            .file_viewer_option {
                background-color: #f6eacb;
                text-align: center;
                border-radius: 3px;
                flex: 1;
                color: #5b3f2d;
                font-weight: bold;
                cursor: default;
                user-select: none;
                transition-duration: 125ms;
                min-width: 90px;

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

.file_transferring_container {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    background-color: #0003;
    backdrop-filter: blur(2px);

    .file_transferring {
        min-width: 400px;
        min-height: 80px;
        background-color: #edb186;
        display: flex;
        flex-flow: column;
        padding: 8px;
        border-radius: 4px;
        color: #583d33;
        font-weight: bold;

        .file_transferring_progress_bar {
            margin: 20px 4px;
            background-color: white;
            height: 16px;
            border-radius: 7px;
            overflow: hidden;

            .file_transferring_progress_rate {
                background-image: linear-gradient(120deg,
                        $progress_bar_clr1 0%,
                        $progress_bar_clr1 33%,
                        $progress_bar_clr2 33%,
                        $progress_bar_clr2 66%,
                        $progress_bar_clr1 66%,
                        $progress_bar_clr1 100%);
                height: 100%;
                background-size: 15px 100%;
                transition-duration: 125ms;
            }

            &:after {
                content: attr(data-rate);
            }
        }
    }
}
</style>
