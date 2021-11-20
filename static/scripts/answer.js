function answer0() {
    const roomId = document.getElementById('room-id').textContent
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=-50`
}

function answer1() {
    const roomId = document.getElementById('room-id').textContent
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=50`
}

function answer2() {
    const roomId = document.getElementById('room-id').textContent
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=100`
}
