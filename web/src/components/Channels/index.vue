<script lang="ts">
import { defineComponent, onBeforeUnmount, onMounted, ref } from 'vue';
import Client, { ClientEvent, Channel } from "@Src/client";

export default defineComponent({
    props: {
        client: {
            type: Client,
            required: true
        }
    },
    setup: ({ client }) => {

        const channels = ref<Map<string, Channel>>(new Map());

        const AppendedChannelHandler = (e: ClientEvent<"ChannelAppended">) => {
            const ch = e.detail;
            channels.value.set(ch.id, ch);
        };

        const RemovedChannelHandler = (e: ClientEvent<"ChannelRemoved">) => {
            const ch = e.detail;
            channels.value.delete(ch.id);
        };

        const ChannelStatusChangedHandler = (e: ClientEvent<"ChannelStatusChanged">) => {
            const ch = e.detail;
            channels.value.set(ch.id, ch);
        };

        const AppendedMessageHandler = (e: ClientEvent<"MessageAppended">) => {
            const { channel } = e.detail;
            channels.value.delete(channel.id);
            channels.value.set(channel.id, channel);
        };

        onMounted(() => {
            client.Channels.forEach(channel => {
                channels.value.set(channel.id, channel);
            });

            client.on("ChannelAppended", AppendedChannelHandler);
            client.on("ChannelRemoved", RemovedChannelHandler);
            client.on('ChannelStatusChanged', ChannelStatusChangedHandler);
            client.on("MessageAppended", AppendedMessageHandler);
        });

        onBeforeUnmount(() => {
            client.off("ChannelAppended", AppendedChannelHandler);
            client.off("ChannelRemoved", RemovedChannelHandler);
            client.off('ChannelStatusChanged', ChannelStatusChangedHandler);
            client.off("MessageAppended", AppendedMessageHandler);
        });

        const SelectChannel = (e: MouseEvent) => {
            const element = e.target as HTMLDivElement;
            client.SetFocusChannel(element.dataset.channel);
        };

        return {
            channels,
            SelectChannel
        };
    }
});

</script>

<template>
    <div class="channels">
        <div class="channel" v-for="({ id, target }) in Array.from(channels.values()).reverse()" :key="id">
            <div class="channel_name" :data-channel="id" @click.stop="SelectChannel">
                {{ target.name || target.id }}
            </div>
            <!--
            <div class="remove_channel">
                <ion-icon name="close"></ion-icon>
            </div>
            -->
        </div>
    </div>
</template>

<style lang="scss">
.channels {
    width: 208px;
    background-color: #654F3A;

    .channel {
        position: relative;
        line-height: 2.4rem;
        padding: 0 10px;
        background-color: transparent;
        transition-duration: 125ms;
        display: flex;
        flex-flow: row;

        .channel_name {
            font-size: 1rem;
            flex: 1;
            text-align: left;
            overflow: hidden;
            text-wrap: nowrap;
            text-overflow: ellipsis;
            cursor: default;
            user-select: none;
            -webkit-user-select: none;
        }

        .remove_channel {
            font-size: 2rem;
            display: none;
            align-items: center;
        }

        &:hover {
            background-color: #0003;

            .remove_channel {
                display: flex;
            }
        }

        &:active {

            background-color: #0005;
        }
    }
}
</style>