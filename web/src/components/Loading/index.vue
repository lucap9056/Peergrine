<script lang="ts">
import { defineComponent, onUnmounted, ref } from 'vue';
import { loadingManager, LoadingManagerEvent } from "./";

export default defineComponent({
    setup: () => {
        const loading = ref<boolean>(loadingManager.IsLoading);

        const StateChangedHandler = (e: LoadingManagerEvent<"StateChanged">) => {
            loading.value = e.detail;
        }

        loadingManager.on("StateChanged", StateChangedHandler);

        onUnmounted(() => {
            loadingManager.off("StateChanged", StateChangedHandler);
        });

        return {
            loading
        }
    }
})
</script>

<template>
    <template v-if="loading">
        <div class="loader_container">
            <div class="loader"></div>
        </div>
    </template>
</template>

<style lang="scss" scoped>
.loader_container {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    background-color: #1113;


    .loader {
        position: relative;
        display: flex;
        align-items: center;
        justify-content: center;
        width: 100%;
        max-width: 6rem;
        margin-top: 3rem;
        margin-bottom: 3rem;

        &:before,
        &:after {
            content: "";
            position: absolute;
            border-radius: 50%;
            animation: pulsOut 1.8s ease-in-out infinite;
            filter: drop-shadow(0 0 1rem rgba(255, 255, 255, 0.75));
        }

        &:before {
            width: 100%;
            padding-bottom: 100%;
            box-shadow: inset 0 0 0 1rem #fff;
            animation-name: pulsIn;
        }

        &:after {
            width: calc(100% - 2rem);
            padding-bottom: calc(100% - 2rem);
            box-shadow: 0 0 0 0 #fff;
        }
    }
}

@keyframes pulsIn {
    0% {
        box-shadow: inset 0 0 0 1rem #fff;
        opacity: 1;
    }

    50%,
    100% {
        box-shadow: inset 0 0 0 0 #fff;
        opacity: 0;
    }
}

@keyframes pulsOut {

    0%,
    50% {
        box-shadow: 0 0 0 0 #fff;
        opacity: 0;
    }

    100% {
        box-shadow: 0 0 0 1rem #fff;
        opacity: 1;
    }
}
</style>