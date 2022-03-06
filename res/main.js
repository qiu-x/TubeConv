function sleep(ms) {
	return new Promise(resolve => setTimeout(resolve, ms));
}

function post(url, data) {
	return fetch(url, {method: "POST", redirect: "follow", body: JSON.stringify(data)})
}

function log(text) {
	let caller = arguments.callee.caller
	if (caller != null) {
		console.log(caller.name.toString() + ": " + text)
	} else {
		console.log("main: " + text)
	}
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
		});
	}
}

function makeLocalSpinner() {
	let spn_clone = document.getElementById("spinner").cloneNode(true);
	spn_clone.style.position = "absolute";
	spn_clone.style.top = "0px";
	spn_clone.style.marginTop = "0px";
	spn_clone.style.marginLeft = "auto";
	spn_clone.style.marginRight = "auto";
	spn_clone.style.left = "0px";
	spn_clone.style.right = "0px";
	spn_clone.style.zIndex = 1;
	spn_clone.classList.add("resp-spinner");
	let spn = {
		spinner: spn_clone,
		show: function() {
			let spi = this.spinner;
			spi.style.display = "inline-block";
			window.setTimeout(function(){
				spi.style.opacity = 1;
			},0);
		},
		hide: function(func) {
			let spi = this.spinner;
			spi.style.opacity = 0;
			spi.addEventListener("transitionend", () => {
				spi.style.display = "none";
				if (func) func();
			});
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
	document.activeElement.blur();
	await clearResoult();
	txt = document.getElementById("search-input").value;
	await checkLink(txt);
	await showResoult();
}

async function showResoult() {
	let search = document.getElementById("search");
	if (!showResoult.transformed) {
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
	if (!showResoult.transformed) {
		search.addEventListener("transitionend", show);
	} else {
		vids.item(0).addEventListener("transitionend", show);
	}
	showResoult.transformed = true;
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
	});
}

async function checkLink(txt) {
	const resp = await post("/req", {
		request: "check-link",
		link: txt
	});
	const json = await resp.json();
	switch(json.type) {
	case "title":
		await queryVideo(txt);
		break;
	case "link":
		// videoInfo(txt);
		break;
	default:
		log("bad response")
		break;
	}
}

async function queryVideo(title) {
	const resp = await post("/req", {
		request: "query",
		text: title
	});
	const json = await resp.json();
	const template = document.getElementById("video-template");
	const json_videos = json.videos;
	console.log(json);
	for (let i = 0; i < json_videos.length; i++) {
		vid = template.cloneNode(true);
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
		dl_button = vid.getElementsByClassName("download-button").item(0);
		dl_button.onclick = () => {videoInfo(i, json_videos[i].link);};
		dl_button.id = json_videos[i].link;
		document.getElementById("main-view").appendChild(vid);
	}
}

async function videoInfo(index, link) {
	let elem = document.getElementsByClassName("video").item(index);

	// Animate background
	let bg = elem.getElementsByClassName("alt-bg").item(0)
	bg.style.width = "100%";
	let dl_button = elem.getElementsByClassName("download-button").item(0);
	dl_button.classList.add("final")
	let arrow = dl_button.getElementsByTagName("img").item(0);
	arrow.classList.add("final");

	let info = elem.querySelector(".video-info")

	// Make spinner
	let spn = makeLocalSpinner();
	info.appendChild(spn.spinner);
	bg.addEventListener("transitionend", () => spn.show());

	// Send request
	const resp = await post("/req", {
		request: "video-info",
		link: link
	});
	const json = await resp.json();
	const template = document.getElementById("video-template");
	const json_videos = json.videos;

	spn.hide(() => showInfo(info,json));

	makeDownloadButton(elem);
}

function selItem(item) {
	let items = Array.from(item.parentNode.getElementsByClassName("item"));
	for (let i = 0; i < items.length; i++) {
		items[i].classList.remove("selected");
	}
	item.classList.add("selected")
}

function showInfo(infoElem, json) {
	// Store the current element height
	let old_h = getComputedStyle(infoElem).height;
	infoElem.style.height = old_h;
	// Load element template
	infoElem.innerHTML = document.getElementById("video-info-temp").innerHTML;
	// Restore the old height (makes the animation work)
	infoElem.style.opacity = 0;

	// Helper function for filling the menu items
	let assign_text = (menu, arr) => {
		let item_empty = menu.getElementsByClassName("item").item(0);
		for (let i = 0; i < arr.length; i++) {
			let item = item_empty.cloneNode(true);
			item.innerText = arr[i];
			menu.appendChild(item);
		}
		item_empty.remove();
	}

	// Fill the menus
	let menus = Array.from(infoElem.getElementsByClassName("scrollmenu"));
	let menu_info = [
		json.video_quality.reverse(),
		json.audio_quality.reverse(),
		["mp4", "webm", "mp3", "ogg"]
	]
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

async function makeDownloadButton(vidElem) {
	dl_btn = vidElem.getElementsByClassName("download-button").item(0);
	dl_btn.onclick = () => {
		let l_dl_btn = vidElem.getElementsByClassName("download-button").item(0);
		let l_menus = Array.from(vidElem.getElementsByClassName("scrollmenu"));
		downloadVideo(l_menus, l_dl_btn.id);
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
	aud_q = getSel(menus[1].getElementsByClassName("item")).innerText.trim();
	fmt = getSel(menus[2].getElementsByClassName("item")).innerText.trim();
	let req_cont = {
		request: "download",
		link: link,
		format: fmt
	}
	req_cont["video-quality"] = String(vid_q);
	req_cont["audio-quality"] = Number(+(aud_q));
	console.log(req_cont);
	const resp = await post("/req", req_cont);
	const json = await resp.json();
	window.location = json.file;
}
