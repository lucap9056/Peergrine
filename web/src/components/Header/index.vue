<script lang="ts">
import { defineComponent, onMounted, onUnmounted, ref } from 'vue';
import ClientProfile, { ClientProfileEvent } from "@Src/client/ClientProfile";

export default defineComponent({
    props: {
        profile: {
            type: ClientProfile,
            required: true
        }
    },
    setup: ({ profile }) => {
        const userId = ref(profile.ClientId);
        const userName = ref(profile.ClientName);
        const userNameEditing = ref(false);

        const ClientNameChangedHandler = (e: ClientProfileEvent<"ClientNameChanged">) => {
            userName.value = e.detail;
        }

        onMounted(() => {
            profile.on("ClientNameChanged", ClientNameChangedHandler);
        });

        onUnmounted(() => {
            profile.off("ClientNameChanged", ClientNameChangedHandler);
        });

        const HandleSaveUserName = (e: KeyboardEvent) => {
            if (e.key !== "Enter") {
                return;
            }
            e.preventDefault();
            const input = e.target as HTMLInputElement;

            profile.UpdateClientName(input.value);

            userNameEditing.value = false;
        }

        const HandleUserNameEdit = () => {
            userNameEditing.value = true;
        }

        return {
            userId,
            userName,
            userNameEditing,
            HandleSaveUserName,
            HandleUserNameEdit,
        }
    }
})
</script>

<template>
    <div class="header">
        <input v-if="userNameEditing" type="text" class="user_name_edit_input" :placeholder="userId" v-model="userName"
            @keydown.stop="HandleSaveUserName" />
        <template v-else>

            <div class="user_name">{{ userName || userId }}</div>
            <div class="user_name_edit" @click.stop="HandleUserNameEdit">
                <ion-icon name="create"></ion-icon>
            </div>
        </template>
    </div>
</template>

<style lang="scss" scoped>
.header {
    display: flex;
    height: 64px;
    background: linear-gradient(80deg, #563D2B 10rem, #654231 100%);
    box-shadow: 0 0 10px 1px #4B4237;
    flex-flow: row;
    justify-content: left;
    align-items: center;
    padding-left: 20px;
    gap: 5px;


    .user_name_edit_input {
        background-color: #F6EACB;
        border: none;
        outline: none;
        height: 36px;
        width: 240px;
        text-align: center;
        color: #413A31;
        font-size: 1.4rem;
        border-radius: 3px;
    }

    .user_name {
        font-size: 1.4rem;
    }

    .user_name_edit {
        display: flex;
        justify-content: center;
        align-content: center;
        font-size: 2rem;
    }
}
</style>