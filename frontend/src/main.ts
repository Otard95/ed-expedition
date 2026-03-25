import "./style.css";
import App from "./App.svelte";

const loadingScreen = document.getElementById("loading-screen");
const style = document.getElementById("default-styles");
console.log(style);

const app = new App({
	target: document.getElementById("app"),
});

if (loadingScreen) {
	loadingScreen.remove();
}
if (style) {
	style.remove();
}

export default app;
