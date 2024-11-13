<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from 'vue';
import { ROUTE_PATHS } from "@Src/router";  // Constants should be in UPPER_SNAKE_CASE
import { useRoute } from 'vue-router';
import Client, { ClientEvent } from '@Src/client';

export default defineComponent({
    props: {
        client: {
            type: Object as () => Client,
            required: true
        }
    },
    setup: ({ client }) => {
        const route = useRoute();

        const channelNotify = ref<boolean>(false);

        const MessageAppendedHandler = (e: ClientEvent<"MessageAppended">) => {
            if (client.FocusedChannel && client.FocusedChannel.id === e.detail.channel.id) {
                return;
            }

            if (route.path !== ROUTE_PATHS.CHANNELS) {
                channelNotify.value = true;
            }
        };

        onMounted(() => {
            client.on("MessageAppended", MessageAppendedHandler);
        })

        onUnmounted(() => {
            client.off("MessageAppended", MessageAppendedHandler);
        });

        const RemoveChannelNotification = () => {
            channelNotify.value = false;
        }

        return {
            channelNotify,
            RemoveChannelNotification,
            ROUTE_PATHS,
        }
    }
})
</script>


<template>

    <div class="side_options">

        <router-link :to="ROUTE_PATHS.NONE">
            <div class="side_option">
                <div class="side_option_icon">
                    <img src="/favicon.ico" />
                </div>
            </div>
        </router-link>

        <router-link :to="ROUTE_PATHS.CHANNELS">
            <div class="side_option" data-alt="channels" :data-notify="channelNotify"
                @click="RemoveChannelNotification">
            </div>
        </router-link>

        <router-link :to="ROUTE_PATHS.RTC_INVITE">
            <div class="side_option" data-alt="rtc">
            </div>
        </router-link>

        <router-link :to="ROUTE_PATHS.RELAY_INVITE">
            <div class="side_option" data-alt="relay">
            </div>
        </router-link>

        <router-link :to="ROUTE_PATHS.SETTINGS">
            <div class="side_option" data-alt="settings">
            </div>
        </router-link>

    </div>

</template>

<style lang="scss" scoped>
.side_options {
    width: 56px;
    display: flex;
    flex-flow: column;
    background: linear-gradient(10deg, #66452E 0%, #4d2714 100%);
    padding: 8px;
    gap: 8px;
}

.side_option {
    position: relative;
    aspect-ratio: 1;
    border: solid 4px #C4A78E;
    border-radius: 100%;
    background-color: #E4B88A;
    transition-duration: 125ms;

    .side_option_icon {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        overflow: hidden;

        img {
            width: 100%;
            height: 100%;
            object-fit: contain;
        }
    }

    &::after {
        content: attr(data-alt);
        position: absolute;
        bottom: 0;
        left: 100%;
        background-color: #2a1106;
        padding: 0 5px;
        line-height: 1.6rem;
        border-radius: 4px;
        color: white;
        pointer-events: none;
        opacity: 0;
        transition-duration: 125ms;
        z-index: 1;
    }

    &[data-notify="true"] {
        animation: notifyAnimation 2s ease-in-out infinite;
    }

    @keyframes notifyAnimation {

        0%,
        100% {
            border-color: #C4A78E;
            box-shadow: 0 0 5px 1px transparent, 0 0 10px 1px transparent inset;
        }

        50% {
            border-color: #f82d17;
            box-shadow: 0 0 5px 1px #da3d2c, 0 0 10px 1px #eb8e85 inset;
        }
    }

    &:hover {
        border-color: #9e8570;
        background-color: #b6916a;

        &::after {
            opacity: 1;
        }
    }

}
</style>