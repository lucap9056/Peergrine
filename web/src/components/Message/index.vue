<script lang="ts">
import { defineComponent, PropType } from 'vue';

import { Message } from "@Src/storage/message";
import FileManager from "@Src/storage/files";

export default defineComponent({
    props: {
        message: {
            type: Object as () => Message,
            required: true
        },
        own: {
            type: Boolean as PropType<boolean>,
            required: true
        },
        received: {
            type: Boolean as PropType<boolean>,
            required: true
        },
        fileManager: {
            type: Object as () => FileManager,
            required: true
        },
    },
    setup: (props) => {
        const { message, own, fileManager } = props;

        const FormatFileSize = (bytes: number) => {
            const units = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            let unitIndex = 0;

            while (bytes >= 1024 && unitIndex < units.length - 1) {
                bytes /= 1024;
                unitIndex++;
            }

            return `${bytes.toFixed(2)} ${units[unitIndex]}`;
        }

        const HandleClickFile = () => {
            if (own) return;
            if (message.type === "FILE") {
                const fileInfo = message.content;
                fileManager.DisplayFile(fileInfo.id);
            }
        }

        return {
            ...message,
            FormatFileSize,
            HandleClickFile
        }
    }
});
</script>

<template>

    <div class="chat_message" :key="message.id" :data-type="type" :data-own="own" :data-received="received">

        <template v-if="type === 'TEXT'">
            <span class="chat_message_text">
                {{ content.text }}
            </span>
        </template>

        <template v-if="type === 'FILE'">
            <div class="chat_message_file" @click="HandleClickFile">

                <div class="chat_message_file_icon">
                    <ion-icon name="document"></ion-icon>
                </div>

                <div class="chat_message_file_info">

                    <div class="chat_message_file_name">
                        {{ content.name }}
                    </div>

                    <div class="chat_message_file_size">
                        {{ FormatFileSize(content.size) }}
                    </div>

                </div>

            </div>
        </template>

    </div>
</template>

<style scoped>
.chat_message {
    position: relative;
    padding: 3px 10px;

    .chat_message_text {
        padding: 0 20px;
        background-color: #824c37;
        border-radius: 5px;
    }

    .chat_message_file {
        background-color: #824c37;
        padding: 0.5rem;
        border-radius: 8px;
        display: flex;
        flex-flow: row;
        transition-duration: 500ms;

        &:hover {
            transform: scale(1.02);
            box-shadow: -2px 3px 4px 1px #694b3e;
            transition-duration: 125ms;
        }


        &:active {
            transform: scale(0.95);
            box-shadow: -1px 3px 4px #bd8c75b0, -1px 3px 8px #3e281f inset, 1px -3px 5px 1px #a16e5a inset;
            transition-duration: 125ms;
        }

        .chat_message_file_icon {
            width: 3rem;
            aspect-ratio: 1;
            font-size: 3rem;
            overflow: hidden;
            display: flex;
            justify-content: center;
            align-items: center;
        }

        .chat_message_file_info {
            text-align: left;
            cursor: default;

            .chat_message_file_name {
                font-size: 1rem;
            }

            .chat_message_file_size {
                font-size: 0.8rem;
            }
        }
    }

    &:hover {
        background: #625442;
    }

    &[data-type="TEXT"] {
        text-align: left;

        &[data-own="true"] {
            text-align: right;
        }
    }

    &[data-type="FILE"] {
        display: flex;
        flex-flow: row;
        justify-content: start;

        &[data-own="true"] {
            justify-content: end;
        }
    }
}
</style>