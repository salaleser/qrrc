function answer() {
    const roomId = document.getElementsByClassName('title')[0].id
    document.location.href = `/spotify/nonamegamego/answer?room=${roomId}`
}

function answer0() {
    const roomId = document.getElementsByClassName('title')[0].id
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=-50`
}

function answer1() {
    const roomId = document.getElementsByClassName('title')[0].id
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=50`
}

function answer2() {
    const roomId = document.getElementsByClassName('title')[0].id
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&correct=100`
}

function exit() {
    document.location.href = `/`
}

function play() {
    document.location.href = `/spotify/nonamegamego/lobby`
}

function deleteRoom() {
    const roomId = document.getElementsByClassName('title')[0].id
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&status=delete`
}

function restartRoom() {
    const roomId = document.getElementsByClassName('title')[0].id
    document.location.href = `/spotify/nonamegamego/room?room=${roomId}&status=start`
}

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

    const roomId = document.getElementsByClassName('title')[0].id

    document.location.href = `/spotify/nonamegamego/room?status=start&player_names=${playerNames.join(',')}&room=${roomId}`
}

function setPlaylist() {
    const playerNames = getPlayerNames()

    const v = document.getElementById('playlist').value
    if (v === '-') {
        v = ''
    }
    
    const roomId = document.getElementsByClassName('title')[0].id
    document.location.href = `/spotify/nonamegamego/setup?playlist=${v}&room=${roomId}&player_names=${playerNames.join(',')}`
}

const newRoom = document.getElementById('new-room');
newRoom.addEventListener('input', () => {
    newRoom.setCustomValidity('');
    newRoom.checkValidity();
});
newRoom.addEventListener('invalid', () => {
    if(newRoom.value === '') {
        newRoom.setCustomValidity('Введите имя комнаты');
    } else {
        newRoom.setCustomValidity('Имя комнаты не должно содержать символы, ' +
            'отличные от маленьких латинских букв, цифр и знака подчеркивания');
    }
});

function addRoom() {
    const roomId = newRoom.value

    const found = roomId.match(/[^a-z0-9_]/g)
    if (found) {
        alert('Имя комнаты не должно содержать символы, ' +
            'отличные от маленьких латинских букв, цифр и знака подчеркивания')
        return
    }

    const c = document.getElementsByClassName('room')
    for (let i = 0; i < c.length; i++) {
        if (c.item(i).textContent === roomId) {
            alert('Комната с таким именем уже существует')
            return
        }
    }

    const clientId = document.getElementById('client-id').value
    const clientSecret = document.getElementById('client-secret').value
    if (clientId === '' || clientSecret === '') {
        alert('Client ID и/или Client Secret пустые')
        return
    }
    
    document.location.href = `/spotify/nonamegamego/setup?room=${roomId}&client_id=${clientId}&client_secret=${clientSecret}`
}
