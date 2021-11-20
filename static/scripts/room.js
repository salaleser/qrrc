function answer() {
    const roomId = document.getElementById('room-id').textContent
    document.location.href = `/spotify/nonamegamego/answer?room=${roomId}`
}

function deleteRoom() {
    const roomId = document.getElementById('room-id').textContent
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&status=delete`
}

function restartRoom() {
    const roomId = document.getElementById('room-id').textContent
    const playerNames = []
    const c = document.getElementsByClassName('player-name')
    for (let i = 0; i < c.length; i++) {
        playerNames.push(c.item(i).textContent)
    }
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&status=start&player_names=${playerNames.join(',')}`
}

const timeLeft = parseInt(document.getElementsByClassName('timer')[0].id)
const dest = new Date().getTime() + timeLeft
let x = setInterval(function() {
    const timer = dest - new Date().getTime()
    document.getElementsByClassName('timer')[0].innerHTML = `Осталось: ${Math.floor(timer / 1000)}`
    if (timer <= 0) {
        clearInterval(x)
        answer()
    }
}, 1000)
