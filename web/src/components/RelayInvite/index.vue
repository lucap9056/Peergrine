<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { ROUTE_PATHS } from "@Src/router"
import Client from '@Src/client';
import RelayManager, { RelayManagerEvent } from "@Src/client/RelayManager";

import { Loading, loadingManager } from "@Components/Loading";

// QRCode class used for generating QR codes
declare class QRCode {
    constructor(element: HTMLElement | string, options: {
        text: string
        width?: number
        height?: number
        colorDark?: string
        colorLight?: string
    });
}

export default defineComponent({
    props: {
        client: {
            type: Object as () => Client,
            required: true
        }
    },
    setup(props) {
        const { Relay } = props.client;
        const router = useRouter();

        const enabled = ref<boolean>(Relay.Enabled);
        const linkCode = ref<string>();
        const qrCodeRef = ref<HTMLElement>();
        const codeValidityDuration = ref<string>();
        const loadingState: { sse?: Loading, qrcode?: Loading, appendUser?: Loading } = {};

        const FormatDuration = (seconds: number): string => {
            const minutes = Math.floor(seconds / 60);
            const secs = seconds % 60;
            return `${minutes}:${secs.toString().padStart(2, "0")}`;
        };

        const GenerateQRCode = (text: string) => {
            const canvas = qrCodeRef.value;
            if (canvas) {
                while (canvas.firstChild) {
                    canvas.removeChild(canvas.firstChild);
                }

                new QRCode(canvas, {
                    text,
                    width: 256,
                    height: 256,
                    colorDark: "#4B4237",
                    colorLight: "transparent"
                });
            }
        };

        const UserAppendedHandler = (_: RelayManagerEvent<"UserAppended">) => {
            if (loadingState.appendUser) {
                loadingState.appendUser.Remove();
                delete loadingState.appendUser;
            }
        }

        const StateChangedHandler = (e: RelayManagerEvent<"StateChanged">) => {
            switch (e.detail) {
                case RelayManager.Status.CONNECTED: {
                    if (loadingState.sse) {
                        loadingState.sse.Remove();
                    }
                    break;
                }
            }
        }

        const DurationUpdatedHandler = (e: RelayManagerEvent<"LinkCodeDurationUpdated">) => {
            codeValidityDuration.value = e.detail > 0 ? FormatDuration(e.detail) : "";
        };

        const ErrorOccurredHandler = (_: RelayManagerEvent<"ErrorOccurred">) => {
            if (loadingState.appendUser) {
                loadingState.appendUser.Remove();
                delete loadingState.appendUser;
            }
        }

        onMounted(() => {
            const code = Relay.LinkCode;
            if (code) {
                GenerateQRCode(code.link_code);
            } else {
                codeValidityDuration.value = "";
            }
            Relay.on("UserAppended", UserAppendedHandler);
            Relay.on("StateChanged", StateChangedHandler);
            Relay.on("LinkCodeDurationUpdated", DurationUpdatedHandler);
            Relay.on("ErrorOccurred", ErrorOccurredHandler);
        });

        onUnmounted(() => {
            Relay.off("UserAppended", UserAppendedHandler);
            Relay.off("StateChanged", StateChangedHandler);
            Relay.off("LinkCodeDurationUpdated", DurationUpdatedHandler);
            Relay.off("ErrorOccurred", ErrorOccurredHandler);

            if (loadingState.qrcode) {
                loadingState.qrcode.Remove();
                delete loadingState.qrcode;
            }
            if (loadingState.appendUser) {
                loadingState.appendUser.Remove();
                delete loadingState.appendUser;
            }
        });

        const HandleEnable = async () => {
            const value = !enabled.value;
            enabled.value = value;

            if (value) {
                const loading = loadingManager.Add();
                Relay.Enable().catch(() => {
                    enabled.value = false;
                    loading.Remove();
                });
                loadingState.sse = loading;
            } else {
                Relay.Disable();
                if (linkCode.value) {
                    Relay.RemoveLinkCode();
                }
            }
        }

        const HandleUserAppend = (e: KeyboardEvent) => {
            const input = e.target as HTMLInputElement;
            if (e.key !== "Enter") return;

            loadingState.appendUser = loadingManager.Add();
            Relay.GetUserSession(input.value).then(() => {
                router.push(ROUTE_PATHS.CHANNELS);
            }).catch(() => {
                if (loadingState.appendUser) {
                    loadingState.appendUser.Remove();
                }
            });
            input.value = "";
        }

        const HandleRequestLinkCode = () => {

            loadingState.qrcode = loadingManager.Add();

            Relay.RequestLinkCode().then((code) => {
                GenerateQRCode(code);
                linkCode.value = code;
                if (loadingState.qrcode) {
                    loadingState.qrcode.Remove();
                }
            }).catch(() => {
                if (loadingState.qrcode) {
                    loadingState.qrcode.Remove();
                }
            });
        }

        const HandleRemoveLinkCode = () => {
            Relay.RemoveLinkCode();
        }

        return {
            enabled,
            linkCode,
            codeValidityDuration,
            qrCodeRef,
            HandleEnable,
            HandleUserAppend,
            HandleRequestLinkCode,
            HandleRemoveLinkCode,
        }
    }
})
</script>



<template>
    <div class="relay_invite">

        <div class="relay_invite_enabled" data-label="Connect">
            <div :data-enabled="enabled" @click="HandleEnable"></div>
        </div>

        <div v-if="enabled" class="relay_invite_append_user">
            <input type="text" @keydown="HandleUserAppend" placeholder="LinkCode" />
        </div>

        <div class="relay_invite_link">

            <template v-if="enabled">
                <template v-if="codeValidityDuration === ''">
                    <button class="relay_new_link_code_btn" @click="HandleRequestLinkCode">New LinkCode</button>
                </template>

                <template v-else>
                    <div class="relay_link_expiration_time">{{ codeValidityDuration }}</div>
                    <div class="relay_link_code">{{ linkCode }}</div>
                </template>

            </template>

            <div class="relay_qr_code_display" v-show="enabled && codeValidityDuration !== ''" ref="qrCodeRef"></div>

            <template v-if="enabled && codeValidityDuration !== ''">
                <button class="relay_remove_link_code_btn" @click="HandleRemoveLinkCode">Remove LinkCode</button>
            </template>

        </div>



    </div>
</template>


<style lang="scss" scoped>
.relay_invite {
    width: 208px;
    background-color: #654F3A;

    .relay_invite_append_user {
        height: 36px;
        padding: 15px;
        display: flex;
        flex-direction: column;
        border-bottom: solid 2px #7c7351;

        input {
            background-color: #F6EACB;
            border: none;
            outline: none;
            color: #413A31;
            font-size: 1.4rem;
            border-radius: 3px;
            flex: 1;
            font-weight: bold;
            text-align: center;
        }
    }

    .relay_invite_enabled {
        display: flex;
        flex-flow: row;
        margin-top: 10px;
        justify-content: center;

        &::after {
            content: attr(data-label);
        }

        div {
            position: relative;
            width: 42px;
            background-color: #A07045;
            height: 24px;
            border-radius: 16px;
            border: solid 2px #F6EACB;
            margin: 0 5px;

            &::before {
                content: "";
                position: absolute;
                left: 0;
                width: 20px;
                height: 20px;
                margin: 1px;
                border: 1px solid #F6EACB;
                background-color: #c62d1d;
                border-radius: 10px;
                transition-duration: 250ms;
            }

            &:active::before {
                width: 28px;
            }

            &[data-enabled="true"]::before {
                left: calc(100% - 24px);
                background-color: #63d216;
            }

            &[data-enabled="true"]:active::before {
                left: calc(100% - 32px);
            }
        }

    }
}

.relay_invite_link {
    background: #F6EACB;
    border-style: solid;
    border-width: 0 2px;
    border-color: #736B60;

    .relay_link_expiration_time {
        color: #393025;
        font-size: 1.7rem;
        font-weight: bold;
        font-family: var(--button-font);
        line-height: 2rem;
    }

    .relay_link_code {
        color: #393025;
        font-size: 2rem;
        font-weight: bold;
        font-family: var(--button-font);
        line-height: 2.2rem;
        user-select: all;
    }

    .relay_new_link_code_btn,
    .relay_remove_link_code_btn {
        margin: 10px;
        background: #494138;
        border: none;
        transition-duration: 125ms;

        &:hover {
            background: #42392f;
        }

        &:active {
            background: #352c23;
        }
    }

    .relay_qr_code_display {
        margin: 10px;
    }
}
</style>

<style lang="scss">
.relay_qr_code_display img {
    width: 100%;
    height: 100%;
    object-fit: contain;
}
</style>