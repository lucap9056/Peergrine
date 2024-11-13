<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from 'vue';

import Client, { Channel, User } from '@Src/client';
import ClientProfile from "@Src/client/ClientProfile";

import ChatMessage from '@Src/storage/message';

import Message from "@Components/Message/index.vue";
import FileManager from "@Src/storage/files";
import Chat, { ChatEvent } from "./";

export default defineComponent({
    components: {
        Message
    },
    props: {
        client: {
            type: Object as () => Client,
            required: true
        },
        chat: {
            type: Object as () => Chat,
            required: true
        },
        fileManager: {
            type: Object as () => FileManager,
            required: true
        },
    },
    setup: ({ client, chat, fileManager }) => {

        const clientProfile = ref<ClientProfile>(client.Profile);
        const channel = ref<Channel>();
        const messages = ref<ChatMessage[]>([]);

        const ChannelChangedHandler = (e: ChatEvent<"ChannelChanged">) => {
            channel.value = e.detail;
        }

        const MessagesChangedHandler = (e: ChatEvent<"MessagesUpdated">) => {
            messages.value = e.detail;
        }

        const MessageAppendedHandler = (e: ChatEvent<'MessageAppended'>) => {
            messages.value.push(e.detail);
        }

        onMounted(() => {
            chat.on("ChannelChanged", ChannelChangedHandler);
            chat.on("MessagesUpdated", MessagesChangedHandler);
            chat.on("MessageAppended", MessageAppendedHandler);
        })

        onUnmounted(() => {
            chat.off("ChannelChanged", ChannelChangedHandler);
            chat.off("MessagesUpdated", MessagesChangedHandler);
            chat.off("MessageAppended", MessageAppendedHandler);
        });

        const HandlePasteContent = (e: ClipboardEvent) => {
            const clipboardData = e.clipboardData;

            if (!clipboardData) {
                return;
            }

            const pastedData = clipboardData.getData('Text');
            if (pastedData === "") {
                e.preventDefault();
            }
        }

        const StripHtmlRecursive = (element: HTMLElement) => {
            for (const childNode of element.childNodes) {
                if (childNode.nodeName !== "#text") {
                    const child = childNode as HTMLElement;
                    if (child.childNodes.length > 0) {
                        StripHtmlRecursive(child);
                    }
                    child.outerText = child.innerText;
                }
            }
        }

        const HandleInputContent = (e: Event) => {
            const element = e.target as HTMLDivElement;
            StripHtmlRecursive(element);
        }

        const HandleSelectFile = () => {

            const input = document.createElement("input");
            input.type = "file";
            input.multiple = false;
            input.onchange = () => {
                if (!input.files) return;

                const file = input.files[0];

                chat.SendFileMessage(file);
            }

            input.click();

        }

        const HandleSendMessage = (e: KeyboardEvent) => {
            if (!(e.key === "Enter" && !e.shiftKey)) {
                return;
            }
            e.preventDefault();
            const element = e.target as HTMLDivElement;
            const value = element.innerText;

            if (value !== "") {
                chat.SendTextMessage(value);
                element.innerText = "";
            }
        }

        return {
            client: clientProfile,
            channel,
            messages,
            fileManager,
            HandlePasteContent,
            HandleInputContent,
            HandleSelectFile,
            HandleSendMessage,
            User,
        }
    }
});

</script>

<template>
    <div class="chat_container">

        <template v-if="channel !== undefined && client !== undefined">
            <div class="chat_header">

                <div class="chat_target_name">{{ channel.target.name || channel.target.id }}</div>

            </div>

            <div class="chat">

                <div class="chat_messages">
                    <Message v-for="message in messages" :key="message.id" :message="message.content"
                        :own="message.sender_id === client.ClientId" :received="true" :fileManager>
                    </Message>
                    <div class="chat_message" v-for="{ id } in messages" :key="id"></div>
                </div>

                <div class="chat_input_container">
                    <div class="chat_input_text" contenteditable="true" @keydown.stop="HandleSendMessage"
                        @paste="HandlePasteContent" @input="HandleInputContent">
                    </div>

                    <div v-if="channel.target.manager === User.MANAGER_NAMES.RTC" class="chat_input_file"
                        @click="HandleSelectFile">
                        <ion-icon name="document"></ion-icon>
                    </div>
                </div>
            </div>
        </template>

    </div>
</template>

<style lang="scss" scoped>
.chat_container {
    background: linear-gradient(80deg, #B68E72 20%, #CAA276 100%);
    border-radius: 7px;
    margin: 1rem;
    display: flex;
    flex-flow: column;
    flex: 1;


    .chat_header {
        display: flex;
        flex-flow: row;
        padding: 10px 20px;

        .chat_target_name {
            flex: 1;
            text-align: left;
            font-weight: bold;
            line-height: 1.5rem;
            font-size: 1.3rem;
        }

    }


    .chat {
        margin: 8px;
        padding: 8px;
        border-radius: 7px;
        display: flex;
        flex-flow: column;
        flex: 1;
        background-color: #a97d5d;


        .chat_messages {
            flex: 1;
            display: flex;
            flex-flow: column;
            overflow-y: auto;
            padding: 10px 0;

            &::before {
                content: "";
                flex: 1;
            }

            .chat_message {
                &:hover {
                    background: #0002;
                }

                .chat_message_text {
                    padding: 0 20px;
                    background-color: #9a4b33;
                    border-radius: 5px;
                }
            }
        }

        .chat_input_container {
            background-color: #CDA182;
            border-radius: 5px;
            display: flex;
            flex-flow: row;
            padding: 0 10px;

            &[data-visible="false"] {
                display: none;
            }

            .chat_input_text {
                flex: 1;
                margin: 10px 0 12px 40px;
                font-size: 1rem;
                min-height: 1.1rem;
                line-height: 1.2rem;
                background: transparent;
                text-align: left;
                overflow-x: hidden;
                overflow-wrap: break-word;
                word-wrap: break-word;
                box-sizing: border-box;
                outline: none;
                color: #3c260e;
            }

            .chat_input_file {
                width: 32px;
                height: 32px;
                font-size: 2.5rem;
                display: flex;
                justify-content: center;
                align-items: center;
                margin: 5px;
                border-radius: 5px;
                transition-duration: 125ms;

                &:hover {
                    background-color: #00000054;
                }
            }
        }
    }
}
</style>