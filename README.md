# G-Origin-Tools

**G-Origin-Tools** is a custom extension for G-Earth designed to enhance your in-game experience by providing real-time player tracking, outfit management, and interaction tools through an intuitive GUI. Built using GoEarth, Vue 3, Wails v2, and Vite, this tool allows you to copy outfits, save them for future use, follow other players, and even mimic their chat behavior in real time.

## Features

- **Real-Time Room Tracking**: Displays a list of all users in the room (including yourself) upon entering.
- **Outfit Management**:
  - **Copy Outfit**: Instantly replace your current outfit with the selected user's outfit and reload the room with your new look.
  - **Save Outfit**: Name and save outfits for future use, allowing easy swapping between saved outfits.
  - **Save Your Own Outfit**: Select yourself from the list to save your current outfit for future use.

- **User Interactions**:
  - **Follow User**: When enabled, you can choose a player to follow. Once activated, your avatar will mirror their movements.

  - **Mimic User**: When this feature is enabled, your avatar will mirror the selected user's chat activity, sending the same messages they do. After pressing "Go," anything the player sends in chat will be automatically mimicked by you.


## How It Works

When entering a room, G-Origin-Tools displays a list of all players present, including yourself. From the list, you can:
- **Player Selection**: When you enter a room, the GUI shows a list of all players present.
- **Copy an Outfit**: Select a player from the list and click "Copy Outfit". This will replace your outfit with theirs and reload the room.
- **Save an Outfit**: You can save any outfit from a selected user, or even your own current outfit, by giving it a name. The outfit will be saved to your personal outfit list for future use.
- **Follow/Mimic**: Choose a player to follow or mimic by enabling the respective features and pressing "Go."

## Getting Started

### Prerequisites

- **G-Earth**: Ensure you have G-Earth installed, as this tool is an extension for it.
- **Wails v2**: The cross-platform GUI is built using Wails.
- **Vue 3 + Vite**: This project uses Vue 3 with Vite for fast and easy development.

### Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/JTD420/G-Origin-Tools.git
    ```


2. Navigate to the project directory:
    ```bash
    cd G-Origin-Tools
    ```


3. Install dependencies:
    ```bash
    npm install
    ```


4. Run the application:
    ```bash
    npm run dev
    ```

### Usage

1. Launch the G-Earth client.
2. Load G-Origin-Tools through the G-Earth extension menu.
3. Use the GUI to track players, copy/save outfits, follow users, or mimic their chat behavior.
4. Save outfits to easily switch between them using the "Save Outfit" feature.

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.


## Recommended IDE Setup

- [VS Code](https://code.visualstudio.com/) + [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar)

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests for improvements, features, or bug fixes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.