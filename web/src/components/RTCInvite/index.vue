<script lang="ts">
import { defineComponent, ref, onMounted, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import { ROUTE_PATHS } from "@Src/router";
import Client from '@Src/client';
import { RTCEvent } from "@Src/client/RTCManager";

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
        const { Rtc } = props.client;
        const router = useRouter();

        const linkCode = ref<string>();
        const qrCodeRef = ref<HTMLElement>();
        const codeValidityDuration = ref<string>();

        const loadingState: { qrcode?: Loading, appendUser?: Loading } = {};

        const FormatDuration = (seconds: number): string => {
            const minutes = Math.floor(seconds / 60);
            const secs = seconds % 60;
            return `${minutes}:${secs.toString().padStart(2, "0")}`;
        };

        const GenerateQrCode = (text: string) => {
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

        const OfferReadyReceivedHandler = (e: RTCEvent<"OfferReady">) => {
            const { link_code } = e.detail;
            linkCode.value = link_code;
            GenerateQrCode(link_code);

            if (loadingState.qrcode) {
                loadingState.qrcode.Remove();
                delete loadingState.qrcode;
            }
        };

        const LinkCodeDurationUpdatedHandler = (e: RTCEvent<"LinkCodeDurationUpdate">) => {
            codeValidityDuration.value = e.detail > 0 ? FormatDuration(e.detail) : "";
        };

        const RtcErrorOccurredHandler = (e: RTCEvent<"ErrorOccurred">) => {
            console.log(e);
            if (loadingState.qrcode) {
                loadingState.qrcode.Remove();
                delete loadingState.qrcode;
            }
            if (loadingState.appendUser) {
                loadingState.appendUser.Remove();
                delete loadingState.appendUser;
            }
        };

        onMounted(() => {
            const code = Rtc.LinkCode;
            if (code) {
                GenerateQrCode(code.link_code);
            } else {
                codeValidityDuration.value = "";
            }

            Rtc.on("OfferReady", OfferReadyReceivedHandler);
            Rtc.on("LinkCodeDurationUpdate", LinkCodeDurationUpdatedHandler);
            Rtc.on("ErrorOccurred", RtcErrorOccurredHandler);
        });

        onBeforeUnmount(() => {
            Rtc.off("OfferReady", OfferReadyReceivedHandler);
            Rtc.off("LinkCodeDurationUpdate", LinkCodeDurationUpdatedHandler);
            Rtc.off("ErrorOccurred", RtcErrorOccurredHandler);
            if (loadingState.qrcode) {
                loadingState.qrcode.Remove();
                delete loadingState.qrcode;
            }
            if (loadingState.appendUser) {
                loadingState.appendUser.Remove();
                delete loadingState.appendUser;
            }
        });

        const HandleUserAppend = (e: KeyboardEvent) => {
            const input = e.target as HTMLInputElement;
            if (e.key !== "Enter") return;

            loadingState.appendUser = loadingManager.Add();
            Rtc.Answer(input.value).then(() => {
                router.push(ROUTE_PATHS.CHANNELS);

            });
            input.value = "";
        };

        const HandleCreateNewLinkCode = async () => {
            const loading = loadingManager.Add();
            Rtc.Offer().catch(() => {
                loading.Remove();
            });
            loadingState.qrcode = loading;
        };

        const HandleRemoveLinkCode = () => {
            if (codeValidityDuration.value) {
                Rtc.RemoveLinkCode();
            }
        };

        return {
            linkCode,
            qrCodeRef,
            codeValidityDuration,
            HandleUserAppend,
            HandleCreateNewLinkCode,
            HandleRemoveLinkCode,
            FormatDuration
        };
    }
});
</script>


<template>
    <div class="rtc_invite">
        <div class="rtc_invite_append_user">
            <input type="text" @keydown="HandleUserAppend" placeholder="LinkCode" />
        </div>

        <div class="rtc_invite_link">
            <template v-if="codeValidityDuration === ''">
                <button class="rtc_new_link_code_btn" @click="HandleCreateNewLinkCode">New LinkCode</button>
            </template>

            <template v-else>
                <div class="rtc_link_expiration_time">{{ codeValidityDuration }}</div>
                <div class="rtc_link_code">{{ linkCode }}</div>
            </template>

            <div class="rtc_qr_code_display" v-show="codeValidityDuration" ref="qrCodeRef"></div>

            <template v-if="codeValidityDuration !== ''">
                <button class="rtc_remove_link_code_btn" @click="HandleRemoveLinkCode">Remove LinkCode</button>
            </template>
        </div>
    </div>
</template>

<style lang="scss" scoped>
.rtc_invite {
    width: 208px;
    background-color: #654F3A;

    .rtc_invite_append_user {
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
}

.rtc_invite_link {
    background: #F6EACB;
    border-style: solid;
    border-width: 0 2px;
    border-color: #736B60;

    .rtc_link_expiration_time {
        color: #393025;
        font-size: 1.7rem;
        font-weight: bold;
        font-family: var(--button-font);
        line-height: 2rem;
    }

    .rtc_link_code {
        color: #393025;
        font-size: 2rem;
        font-weight: bold;
        font-family: var(--button-font);
        line-height: 2.2rem;
        user-select: all;
    }

    .rtc_new_link_code_btn,
    .rtc_remove_link_code_btn {
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

    .rtc_qr_code_display {
        margin: 10px;
    }
}
</style>


<style lang="scss">
.rtc_qr_code_display img {
    width: 100%;
    height: 100%;
    object-fit: contain;
}
</style>