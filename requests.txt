REQ:
{
    "request": "check-link",
    "link": "https://youtube.com/theurl",
}
RESP:
{
    "type": "link" // or title
}
--------
REQ:
{
    "request": "query",
    "text": "Crab Rave"
}
RESP:
{
    "videos": [ 
    {
        "title": "Crab Rave", 
        "author": "Noisestorm", 
        "link": "https://youtube.com/theurl", 
        "thumbnail": "https://linktothumbnail"
    },
    {
        "title": "Crab Rave - Rap God", 
        "author": "Noisestorm", 
        "link": "https://youtube.com/theurl", 
        "thumbnail": "https://linktothumbnail"
    }
    ...
    ]
}
--------
REQ:
{
    "request": "video-info",
    "link": "https://youtube.com/theurl"
}
RESP:
{
    "video-quality": [
        "720", "1080", "none"
    ],
    "audio-quality": [
        "120", "240", "360", "480", "none"
    ]
}
--------
REQ:
{
    "request": "download",
    "link": "https://youtube.com/theurl",
    "video-quality": "1080",
    "audio-quality": "360",
    "format": "mp4" // avalible: mp4, mp3, opus, wav, mkv, ogg, webm
}
RESP:
{
    "file": "https://example.com/video"
}

TODO: Download status
