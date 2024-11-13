<script lang="ts">
import { defineComponent, ref } from "vue";

export default defineComponent({
    props: {
        options: {
            type: Array as () => string[],
            required: true
        },
        value: {
            type: String,
            required: true
        }
    },
    emits: ["update"],
    setup(props, { emit }) {
        const isOpen = ref(false);
        const selected = ref(props.value);

        const toggleDropdown = () => {
            isOpen.value = !isOpen.value;
        };

        const selectOption = (option: string) => {
            selected.value = option;
            emit("update", option);
            isOpen.value = false;
        };

        return {
            isOpen,
            selected,
            toggleDropdown,
            selectOption
        };
    }
});
</script>

<template>
    <div class="select">
        <div class="select-selection" @click="toggleDropdown">
            {{ selected }}
        </div>
        <div v-if="isOpen" class="select-menu">
            <div v-for="option in options" :key="option" class="select-item" @click="selectOption(option)">
                {{ option }}
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.select {
    position: relative;
    border: 1px solid #ccc;
    border-radius: 4px;
    margin: 10px;
}

.select-selection {
    padding: 8px;
    cursor: pointer;
    background-color: #d0a279;
    color: #503122;
    font-weight: bold;
}

.select-menu {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background-color: #743d22;
    border: 1px solid #ccc;
    border-radius: 4px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    z-index: 1;
}

.select-item {
    padding: 8px;
    cursor: pointer;

    &:hover {
        background-color: #59311d;
    }
}
</style>
