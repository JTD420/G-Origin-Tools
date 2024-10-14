<template>
  <div class="user-list-section p-4 h-vh pb-16 rounded-lg">
    <h2 class="section-title">User Actions</h2>

    <!-- Follow Toggle Button -->
    <button 
      @click="toggleFollow" 
      class="block w-full p-2 text-lg border rounded my-4"
      :class="followEnabled ? 'bg-green-600 text-white border-green-700' : 'bg-gray-600 text-gray-200 border-gray-700'">
      {{ followEnabled ? 'Follow On' : 'Follow Off' }}
    </button>

    <!-- Mimic Toggle Button -->
    <button 
      @click="toggleMimic" 
      class="block w-full p-2 text-lg border rounded my-4"
      :class="mimicEnabled ? 'bg-green-600 text-white border-green-700' : 'bg-gray-600 text-gray-200 border-gray-700'">
      {{ mimicEnabled ? 'Mimic On' : 'Mimic Off' }}
    </button>

    <!-- Go Button -->
    <button 
      @click="followUser" 
      class="block w-full bg-blue-600 p-2 text-white text-lg border border-blue-700 hover:bg-blue-700 rounded my-4" 
      :disabled="!selectedUser"
      :class="{ 'hidden' : !selectedUser, 'hidden' : followTarget }">
      Go
    </button>

    <button 
    @click="stopFollow" 
    class="block w-full bg-red-600 p-1 text-slate-100 text-lg border border-slate-700 hover:border-slate-900 hover:bg-red-700 rounded my-4" 
    :class="{ 'hidden': !followTarget }" 
    :disabled="!followTarget">
    Stop
  </button>

    <!-- Grid layout for user selection -->
    <div class="max-h-96 overflow-y-auto grid grid-cols-3 gap-2">
      <div 
        v-for="(user, index) in sortedUsers" 
        :key="index" 
        @click="selectUser(user)" 
        :class="['p-2 rounded cursor-pointer', selectedUser === user ? 'bg-gray-700' : 'bg-gray-800 hover:bg-gray-600']">
        {{ user }}
      </div>
    </div>
    
    <!-- Other actions -->
    <button @click="copyOutfit" class="action-button">Copy Outfit</button>
    <button @click="saveOutfit" class="action-button">Save Outfit</button>

    <!-- Load saved outfits section -->
    <div class="my-4">
      <label for="outfitSelect" class="flex mb-2">Select an outfit:</label>
      <select v-model="selectedOutfit" id="outfitSelect" class="w-full">
        <option v-for="(outfit, index) in savedOutfits" :key="index" :value="outfit">{{ outfit }}</option>
      </select>
      <button @click="loadOutfit" class="action-button">Load Outfit</button>
    </div>

    <!-- Activity Logs -->
    <h2 class="section-title">Activity Logs</h2>
    <div id="log" ref="logbox" class="log-section mb-2">
      <div v-for="(msg, index) in log" :key="index">{{ msg }}</div>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      users: [], // List of users in the room
      selectedUser: '', // Currently selected user
      followTarget: null, // Target user being followed
      log: [], // Log for activity
      savedOutfits: [], // List of saved outfits
      selectedOutfit: '', // The outfit selected from the dropdown
      followEnabled: false, // Tracks follow toggle state
      mimicEnabled: false // Tracks mimic toggle state
    };
  },
  computed: {
    // Sort users alphabetically
    sortedUsers() {
      return this.users.slice().sort((a, b) => a.localeCompare(b));
    }
  },
  methods: {
    // Toggle Follow state and call corresponding backend function
    async toggleFollow() {
      this.followEnabled = !this.followEnabled;
      if (this.followEnabled) {
        await window.go.main.App.FollowOn();
        this.addLogMsg('Follow enabled');
      } else {
        await window.go.main.App.FollowOff();
        this.addLogMsg('Follow disabled');
      }
    },
    // Toggle Mimic state and call corresponding backend function
    async toggleMimic() {
      this.mimicEnabled = !this.mimicEnabled;
      if (this.mimicEnabled) {
        await window.go.main.App.MimicOn();
        this.addLogMsg('Mimic enabled');
      } else {
        await window.go.main.App.MimicOff();
        this.addLogMsg('Mimic disabled');
      }
    },
    async copyOutfit() {
      try {
        await window.go.main.App.CopyOutfit(this.selectedUser);
      } catch (error) {
        this.addLogMsg('Error copying outfit');
        console.error(error);
      }
    },
    async saveOutfit() {
      const outfitName = prompt("Enter a name for this outfit:");
      if (outfitName) {
        try {
          await window.go.main.App.SaveOutfitFromUsername(this.selectedUser, outfitName);
          this.addLogMsg(`Outfit saved: ${outfitName}`);
          this.loadSavedOutfits(); // Reload outfits after saving a new one
        } catch (error) {
          this.addLogMsg('Error saving outfit');
          console.error(error);
        }
      }
    },
    async followUser() {
      try {
        await window.go.main.App.FollowUser(this.selectedUser);
        this.followTarget = this.selectedUser;
      } catch (error) {
        this.addLogMsg('Error! Please try again.');
        console.error(error);
      }
    },
    selectUser(user) {
      this.selectedUser = user;
    },
    async loadUsers() {
      try {
        const response = await window.go.main.App.LoadUsers();
        if (response) {
          this.users = response;
        }
      } catch (error) {
        this.addLogMsg('Error fetching users');
        console.error(error);
      }
    },
    async stopFollow() {
      try {
        await window.go.main.App.StopFollowingUser(this.selectedUser);
        this.followTarget = null; // Clear follow target
      } catch (error) {
        this.addLogMsg('Error stopping follow (try toggling on and off)');
        console.error(error);
      }
    },
    async loadOutfit() {
      if (this.selectedOutfit) {
        try {
          await window.go.main.App.LoadAndApplyOutfit(this.selectedOutfit);
        } catch (error) {
          this.addLogMsg('Error loading outfit');
          console.error(error);
        }
      }
    },
    async loadSavedOutfits() {
      try {
        const outfitsConfig = await window.go.main.App.LoadConfig(); // Load the outfits from Go
        // Check if outfitsConfig and outfitsConfig.Outfits exist
        if (outfitsConfig && outfitsConfig.outfits) {
          this.savedOutfits = Object.keys(outfitsConfig.outfits); // Populate the saved outfits
        } else {
          this.savedOutfits = []; // If no outfits are found, set an empty array
          this.addLogMsg('No saved outfits found');
        }
      } catch (error) {
        this.addLogMsg('Error loading saved outfits');
        console.error(error);
      }
    },

    addLogMsg(msg) {
      this.log.push(msg);
      this.$nextTick(() => {
        const logbox = this.$refs.logbox;
        logbox.scrollTop = logbox.scrollHeight;
      });
    },
    scrolldown() {
      this.$nextTick(() => {
        const box = this.$refs.logbox;
        box.scrollTop = box.scrollHeight;
      });
    },
    fetch() {
      this.loadUsers();
    },
  },
  async mounted() {
    this.fetch();
    this.loadSavedOutfits(); // Load saved outfits when component mounts

    window.runtime.EventsOn("logUpdate", (message) => {
      this.log = message.split('\n');
      this.scrolldown();
    });
    window.runtime.EventsOn("usersUpdated", (userList) => {
      this.fetch();
      const usersArray = userList.split(',').map(user => user.trim());
      this.$set(this, 'users', usersArray);
    });
  },
};
</script>

<style scoped>
::-webkit-scrollbar {
    width: 10px;
}

::-webkit-scrollbar-track {
    box-shadow: inset 0 0 10px 10px green;
    border: solid 3px transparent;
}

::-webkit-scrollbar-thumb {
    box-shadow: inset 0 0 10px 10px rgb(58, 3, 3);
    border: solid 3px transparent;
}
/* Add necessary styles for the dropdown and buttons */
.flipped {
  direction: rtl;
}
.user-list-section {
  background-color: #100e0e;
  color: #fff;
}

.section-title {
  font-size: 18px;
  margin-bottom: 10px;
  color: #e0e0e0;
  text-align: center;
}

.form-group {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 10px;
}

label {
  flex: 1;
  font-weight: bold;
  text-align: right;
  margin-right: 10px;
  font-size: 14px;
  color: #c0c0c0;
}

select {
  flex: 2;
  padding: 8px;
  background-color: #2e2e2e;
  border: 1px solid #444;
  border-radius: 4px;
  color: #fff;
  font-size: 14px;
}

.action-button {
  display: block;
  width: 100%;
  padding: 8px;
  background-color: #2f2f2f;
  color: white;
  border: solid .2px #444;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 15px;
  font-size: 14px;
}

.action-button:hover {
  background-color: #1e1e1e;
}

.log-section {
  background-color: #000000;
  padding: 10px;
  border-radius: 4px;
  height: 200px;
  overflow-y: auto;
  font-family: monospace;
  margin-top: 10px;
  color: #00ff00;
  font-size: 13px;
  line-height: 1.4em;
  border: 2px solid #00ff00;
}

.log-section div {
  padding: 2px 0;
}

.update-notice {
  margin-top: 15px;
  padding: 10px;
  background-color: #ffcc00;
  color: #000;
  text-align: center;
  border-radius: 4px;
  font-weight: bold;
}
</style>
