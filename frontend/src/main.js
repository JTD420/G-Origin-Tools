import {createApp} from 'vue'
import App from './App.vue'
import './style.css';

createApp(App).mount('#app')

function UpdateUserList(users) {
    console.log("Users received:", users);
}