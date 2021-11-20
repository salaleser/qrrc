const roomId = getRoomId()

function getPlayerNames() {
    const playerNames = []
    const c = document.getElementsByClassName('name')
    for (let i = 0; i < c.length; i++) {
        const v = c.item(i).value
        if (v !== '') {
            playerNames.push(v)
        }
    }
    return playerNames
}

function start() {
    const playerNames = getPlayerNames()
    if (playerNames.length < 2) {
        alert('Укажите более одного игрока')
        return
    }

    document.location.href = `/spotify/nonamegamego/room?status=start&player_names=${playerNames.join(',')}&room=${roomId}`
}

function setPlaylist() {
    const playerNames = getPlayerNames()

    const v = document.getElementById('playlist').value
    if (v === '-') {
        v = ''
    }
    
    document.location.href = `/spotify/nonamegamego/setup?playlist=${v}&room=${roomId}&player_names=${playerNames.join(',')}`
}
