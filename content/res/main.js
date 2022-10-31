function sleep(ms) {
	return new Promise(resolve => setTimeout(resolve, ms));
}

GlobalTimeout = 200000 // 20sec

// POST request wrapper
async function post(url, ms, data) {
	const controller = new AbortController()
	const signal = controller.signal;
	const promise = fetch(url, { signal: signal, method: "POST", redirect: "follow", body: JSON.stringify(data)});
	if (signal) signal.addEventListener("abort", () => controller.abort());
	const timeout = setTimeout(() => controller.abort(), ms);
	return await promise.finally(() => clearTimeout(timeout));
}

async function ApiCall(req) {
	let resp;
	try {
		resp = await post("/req", GlobalTimeout, req);
	} catch (err) {
		if (err.name === "AbortError") {
			alert("Request timed out.");
		} else if (err.message === "NetworkError when attempting to fetch resource.") {
			alert("Error: Server does not respond. Please check your internet connection.");
		} else if (err.name === "TypeError") {
			alert("AbortSignal.timeout() method is not supported");
		} else {
			// A network error, or some other problem.
			alert("Error: Server does not respond. Please check your internet connection.");
		}
		return err
	}
	if (!resp.ok) {
		let msg;
		if (resp.status == 429) {
			msg = "Request limit exceeded. Please wait for a moment.";
		} else {
			msg = `Error: ${resp.status}`;
		}
		return Error(msg);
	}
	return resp
}

function log(text) {
	let caller = arguments.callee.caller
	if (caller != null) {
		console.log(caller.name.toString() + ": ", text)
	} else {
		console.log("main: ", text)
	}
}

function removeListeners(elem) {
	const oldElement = elem
	const newElement = elem.cloneNode(true);
	oldElement.parentNode.replaceChild(newElement, oldElement);
}

function spinner(show, func) {
	let spinner = document.getElementById("spinner");
	if (show) {
		spinner.style.display = "inline-block";
		window.setTimeout(function(){
			spinner.style.opacity = 1;
		},0);
	} else {
		spinner.style.opacity = 0;
		spinner.addEventListener("transitionend", () => {
			spinner.style.display = "none";
			if (func) func();
		}, {once: true});
	}
}

function newSpinner() {
	let spn_clone = document.getElementById("spinner").cloneNode(true);
	[
		["position", "absolute"],
		["marginTop", "0px"],
		["marginLeft", "auto"],
		["marginRight", "auto"],
		["left", "0px"],
		["right", "0px"],
		["zIndex", 1]
	].map(v => {
		spn_clone.style[v[0]] = v[1];
	});
	spn_clone.classList.add("resp-spinner");
	let spn = {
		elem: spn_clone,
		show: function() {
			let spi = this.elem;
			spi.style.display = "inline-block";
			window.setTimeout(function(){
				spi.style.opacity = 1;
			},0);
		},
		hide: function(func) {
			let spi = this.elem;
			spi.style.opacity = 0;
			spi.addEventListener("transitionend", () => {
				spi.style.display = "none";
				if (func) func();
			}, {once: true});
		}
	}
	return spn;
}

// Execute when the user releases the enter key
document.getElementById("search-input").addEventListener("keydown", event => {
  // Number 13 is the "Enter" key on the keyboard
  if (event.keyCode === 13) {
    // Cancel the default action, if needed
    event.preventDefault();
    // Trigger the same function as onclick
    searchFunc()
  }
});

async function searchFunc() {
	let spn = document.getElementById("spinner");
	removeListeners(spn);
	txt = document.getElementById("search-input").value;
	if (txt == "") {
		return
	}
	document.activeElement.blur();
	await clearResoult();
	ok = await checkLink(txt);
	if (!ok) {
		// window.location.reload();
		spinner(false);
		return;
	}
	await showResult();
}

async function showResult() {
	let search = document.getElementById("search");
	if (!showResult.transformed) {
		search.style.transition = "margin-top 1s";
		search.style.marginTop = 0;
	}
	let vids = document.getElementsByClassName("video");
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.display = "flex";
	}
	let show = () => { spinner(false, () => {
		for (let i = 0; i < vids.length; i++) {
			vids[i].style.visibility = "visible";
			vids[i].style.opacity = 1;
		}
	})}
	if (!showResult.transformed) {
		search.addEventListener("transitionend", show, {once: true});
	} else {
		vids.item(0).addEventListener("transitionend", show, {once: true});
	}
	showResult.transformed = true;
}

async function clearResoult() {
	let vids = document.getElementsByClassName("video");
	if (!vids.item(0)) {
		spinner(true);
		return;
	}
	let main_view = document.getElementById("main-view");
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.opacity = 0;
	}
	let arr = Array.from(vids);
	vids.item(0).addEventListener("transitionend", () => {
		arr.map(x => x.remove())
		spinner(true);
	},{once: true});
}

async function checkLink(txt) {
	let resp = await ApiCall({
		request: "check-link",
		link: txt
	});
	if (resp instanceof Error) {
		return false;
	}
	let ok;
	const json = await resp.json();
	switch(json.type) {
	case "title":
		ok = await queryVideo(txt);
		break;
	case "link":
		ok = await queryVideoByLink(txt);
		break;
	default:
		log("bad response")
		return false;
		break;
	}
	if (!ok) {
		return false;
	}
	return true;
}

async function queryVideoByLink(link) {
	const template = document.getElementById("video-template");
	let vid = template.cloneNode(true);
	vid.id = "video";
	vid.classList.remove("template");
	vid.classList.add("video");
	if (link.substring(0, 8) == "youtu.be") {
		vid.getElementsByClassName("thumb").item(0).src = "https://i.ytimg.com/vi/" +
			link.split("/")[1] + "/hqdefault.jpg";
	} else if (link.substring(0, 16) == "https://youtu.be") {
		vid.getElementsByClassName("thumb").item(0).src = "https://i.ytimg.com/vi/" +
			link.split("/")[3] + "/hqdefault.jpg";
	} else {
		vid.getElementsByClassName("thumb").item(0).src = "https://i.ytimg.com/vi/" +
			link.split("=")[1] + "/hqdefault.jpg";
	}
	let dl_button = vid.getElementsByClassName("download-button").item(0);
	dl_button.id = link;
	dl_button.onclick = () => {videoInfo(0, link);};
	document.getElementById("main-view").appendChild(vid);
	videoInfo(vid, link);
	return true;
}

async function queryVideo(title) {
	let resp = await ApiCall({
		request: "query",
		text: title
	});
	if (resp instanceof Error) {
		return false;
	}
	const json = await resp.json();
	const template = document.getElementById("video-template");
	const json_videos = json.videos;
	for (let i = 0; i < json_videos.length; i++) {
		let vid = template.cloneNode(true);
		vid.id = "video";
		vid.classList.remove("template");
		vid.classList.add("video");
		let san_element = document.createElement('div');
		san_element.innerText = json_videos[i].title;
		vid.getElementsByClassName("video-info").item(0).innerHTML = san_element.innerHTML;
		vid.getElementsByClassName("video-info").item(0).innerHTML += '</br>';
		vid.getElementsByClassName("video-info").item(0).innerHTML += '</br>';
		san_element.innerText = json_videos[i].author;
		vid.getElementsByClassName("video-info").item(0).innerHTML += san_element.innerHTML;
		vid.getElementsByClassName("thumb").item(0).src = json_videos[i].thumbnail;
		let dl_button = vid.getElementsByClassName("download-button").item(0);
		dl_button.onclick = () => {videoInfo(vid, json_videos[i].link);};
		dl_button.id = json_videos[i].link;
		document.getElementById("main-view").appendChild(vid);
	}
	return true;
}

async function videoInfo(element, link) {
	let elem = element;

	// Animate background
	let bg = elem.getElementsByClassName("alt-bg").item(0)
	bg.style.width = "100%";
	bg.style.opacity = 1;
	let dl_button = elem.getElementsByClassName("download-button").item(0);
	dl_button.classList.add("final")
	let arrow = dl_button.getElementsByTagName("img").item(0);
	arrow.classList.add("final");

	let info = elem.querySelector(".video-info")

	// Make spinner
	let spn = newSpinner();
	info.appendChild(spn.elem);
	bg.addEventListener("transitionend", () => spn.show(), {once: true});

	// Send request
	let resp = await ApiCall({
		request: "video-info",
		link: link
	});
	if (resp instanceof Error) {
		window.location.reload();
		return false;
	}
	const json = await resp.json();
	const template = document.getElementById("video-template");
	const json_videos = json.videos;

	let old_h = getComputedStyle(info).height;

	spn.hide(() => showInfo(info, json, old_h));

	newDownloadButton(elem);
}

function selItem(item) {
	let items = Array.from(item.parentNode.getElementsByClassName("item"));
	for (let i = 0; i < items.length; i++) {
		items[i].classList.remove("selected");
	}
	item.classList.add("selected")
}

function showInfo(infoElem, json, old_h) {
	// Load element template
	infoElem.innerHTML = document.getElementById("video-info-temp").innerHTML;
	infoElem.style.opacity = 0;
	infoElem.style.maxHeight = old_h;
	window.setTimeout(function() {
		infoElem.style.maxHeight = "300px";
	},1);

	// Helper function for filling the menu items
	let assign_text = (menu, arr) => {
		let item_empty = menu.getElementsByClassName("item").item(0);
		for (let i = 0; i < arr.length; i++) {
			let item = item_empty.cloneNode(true);
			if (i == 0) {
				item.style.backgroundColor = "#0000004a";
				item.onclick = null;
			}
			if (i == 1) {
				item.classList.add("selected")
			}
			item.innerText = arr[i];
			menu.appendChild(item);
		}
		item_empty.remove();
	}

	// Fill the menus
	let menus = Array.from(infoElem.getElementsByClassName("scrollmenu"));
	json.video_quality.push("Video");
	let menu_info = [
		json.video_quality.reverse(),
		json.audio_quality.map(x => {
			return x+" k/s";
		}),
		["Format", "mp4", "webm", "mp3", "ogg"]
	]
	menu_info[1].unshift("Audio");
	for (let i = 0; i < 3; i++) {
		assign_text(menus[i], menu_info[i]);
	}
	window.setTimeout(function() {
		infoElem.style.opacity = 1;
		infoElem.classList.add("final");
		infoElem.classList.add("showed");
	},0);
	window.setTimeout(function() {
		for (let i = 0; i < menus.length; i++) {
			menus[i].style.width = "30%";
		}
	},0);
}

async function newDownloadButton(vidElem) {
	let dl_btn = vidElem.getElementsByClassName("download-button").item(0);
	dl_btn.onclick = async () => {
		old_onclick = dl_btn.onclick;
		let l_dl_btn = vidElem.getElementsByClassName("download-button").item(0);
		let l_menus = Array.from(vidElem.getElementsByClassName("scrollmenu"));
		let arrow = dl_btn.getElementsByClassName("arrow").item(0);
		arrow.style.filter = "invert(1) brightness(0.5) sepia(1) saturate(0%)"
		await downloadVideo(l_menus, l_dl_btn.id);
		arrow.style.filter = "invert(1) brightness(0.5) sepia(1) saturate(10000%)"
		dl_btn.onclick = old_onclick;
	};
}

async function downloadVideo(menus, link) {
	let getSel = (itm) => {
		for (let i = 0; i < itm.length; i++) {
			if (itm[i].classList.contains("selected")) {
				return itm[i];
			}
		}
		return itm[0];
	};
	vid_q = getSel(menus[0].getElementsByClassName("item")).innerText.trim();
	aud_q = getSel(menus[1].getElementsByClassName("item")).innerText.trim().split(" k/s")[0];
	fmt = getSel(menus[2].getElementsByClassName("item")).innerText.trim();
	let req_cont = {
		request: "download",
		link: link,
		format: fmt
	}
	req_cont["video-quality"] = String(vid_q);
	req_cont["audio-quality"] = Number(+(aud_q));
	let resp = await ApiCall(req_cont);
	if (resp instanceof Error) {
		return false;
	}
	const json = await resp.json();
	window.location = json.file;
}
