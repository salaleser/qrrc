function updateCallbackLink() {
    const roomId = document.getElementById('new-room').value
    const callbackLink = document.getElementById('callback-link')
    callbackLink.value = `https://salaleser.app/callback/${roomId}`
}

function addRoom() {
    const roomId = document.getElementById('new-room').value
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
