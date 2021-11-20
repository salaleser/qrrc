const roomId = getRoomId()

function answer0() {
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=-50`
}

function answer1() {
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=50`
}

function answer2() {
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=100`
}
