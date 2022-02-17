function sleep(ms) {
	return new Promise(resolve => setTimeout(resolve, ms));
}

function post(url, data) {
	return fetch(url, {method: "POST", redirect: "follow", body: JSON.stringify(data)})
}

function log(text) {
	var caller = arguments.callee.caller
	if (caller != null) {
		console.log(caller.name.toString() + ": " + text)
	} else {
		console.log("main: " + text)
	}
}

function moveSearch() {
	search = document.getElementById("search");
	search.style.transform = 'translateY(-100%)';
	vids = document.getElementsByClassName("video")
	for (let i = 0; i < vids.length; i++) {
		vids[i].style.display = "inline-block";
		sleep(i*500).then(() => {
			vids[i].classList.add("showed");
		})
	}
}

async function checkLink(txt) {
	const resp = await post("/api", {
		request: "check-link",
		link: txt
	});
	const json = await resp.json();
	switch(json) {
	case "title":
		queryVideo(txt);
	case "link":
		showVideoInfo(txt);
	default:
		log("bad response")
	}
}

async function queryVideo(title) {
	const resp = await post("/api", {
		request: "query",
		text: title
	});
	const json = await resp.json();
	const template = document.getElementsById("video-template")
	vids = json.videos
	for (i = 0; i < vids.lenght; i++) {
		vid = template.cloneNode(true);
		vid.classList.remove("template");
		vid.classList.append("video");
		vid.getElementById("video-info").html = "Title: "+ vids[i].title;
		vid.getElementById("video-info").html += "Author: "+ vids[i].author;
		vid.getElementById("video-info").html += "Duration: "+ vids[i].duration;
		document.appendChild(vid);
	}
}
