function exit() {
    document.location.href = `/`
}

function play() {
    document.location.href = `/spotify/nonamegamego/lobby`
}

function getRoomId() {
    return document.getElementById('new-room').value
}
