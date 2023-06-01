var invitations = {};
var events = {};
var meetings = {};
var users = {};

function init() {
    getUserInvitations();
}

function getUserInvitations() {
    fetch("/getUsersInvitations", {
        method: "GET"
    })
    .then(response => response.json())
    .then(data => {
        invitations = data;
    })
    .then(getMeetings)
    .then(getEvents)
    .then(getUsers)
    .then(createInvitationCards)
    .catch(error => console.error(error));
}

function getMeetings() {
    const fetchPromises = invitations.map(invitation => {
        const requestBody = { "id": parseInt(invitation.meeting_id) };
        return fetch("/getMeeting", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(requestBody)
        })
        .then(response => response.json())
        .then(data => {
            meetings[data.meeting_id] = data;
        })
        .catch(error => console.error(error));
    });

    // Chain the promises sequentially
    let promiseChain = Promise.resolve();
    fetchPromises.forEach(promise => {
        promiseChain = promiseChain.then(() => promise);
    });

    return promiseChain;
}

function getEvents() {
    return fetch('/getEvents', {
        method: "GET"
    })
    .then(response => response.json())
    .then(data => {
        data.forEach(event => {
            events[event.id] = event;
        });
    })
    .catch(error => console.error(error));
}

function getUsers() {
    const fetchPromises = Object.values(meetings).map(meeting => {
        return fetch(`/getUserByID?userId=${meeting.organizator_id}`, {
            method: "GET"
        })
        .then(response => response.json())
        .then(data => {
            users[data.user_id] = data;
        })
        .catch(error => console.error(error));
    });

    // Chain the promises sequentially
    let promiseChain = Promise.resolve();
    fetchPromises.forEach(promise => {
        promiseChain = promiseChain.then(() => promise);
    });

    return promiseChain;
}

function createInvitationCards() {
    const container = document.getElementById("inboxDiv");
    const container2 = document.getElementById("oldInvitaionsDiv");
    invitations.forEach(invitation => {
        if(invitation.status === "Pending") {
            let meeting = meetings[invitation.meeting_id];
            let event = events[meeting.event_id];
            let userOrg = users[meeting.organizator_id];

            const card = document.createElement("div");
            card.className = "card";

            const organizer = document.createElement("div");
            organizer.className = "organizer";
            organizer.textContent = `Organizer: ${userOrg.username}`;

            const eventName = document.createElement("div");
            eventName.className = "event-name";
            eventName.textContent = `Event: ${event.name}`;

            const date = document.createElement("div");
            date.className = "date";
            date.textContent = `Date: ${event.date}`;

            const buttonsContainer = document.createElement("div");
            buttonsContainer.className = "buttons-container";

            const acceptButton = document.createElement("button");
            acceptButton.textContent = "Accept";
            acceptButton.className = "accept-button";
            acceptButton.addEventListener("click", () => handleAccept(invitation));

            const rejectButton = document.createElement("button");
            rejectButton.textContent = "Reject";
            rejectButton.className = "reject-button";
            rejectButton.addEventListener("click", () => handleReject(invitation));

            buttonsContainer.appendChild(acceptButton);
            buttonsContainer.appendChild(rejectButton);

            card.appendChild(organizer);
            card.appendChild(eventName);
            card.appendChild(date);
            card.appendChild(buttonsContainer);

            container.appendChild(card);
        } else {
            let meeting = meetings[invitation.meeting_id];
            let event = events[meeting.event_id];
            let userOrg = users[meeting.organizator_id];

            const card = document.createElement("div");
            card.className = "card";

            const organizer = document.createElement("div");
            organizer.className = "organizer";
            organizer.textContent = `Organizer: ${userOrg.username}`;

            const eventName = document.createElement("div");
            eventName.className = "event-name";
            eventName.textContent = `Event: ${event.name}`;

            const date = document.createElement("div");
            date.className = "date";
            date.textContent = `Date: ${event.date}`;

            const buttonsContainer = document.createElement("div");
            buttonsContainer.className = "buttons-container";

            const arButton = document.createElement("button");
            arButton.textContent = invitation.status === "Accepted" ? "Reject" : "Accept";
            arButton.className = invitation.status === "Accepted" ? "reject-button" : "accept-button";
            arButton.addEventListener("click", invitation.status === "Accepted" ? () => handleReject(invitation) : () => handleAccept(invitation));

            buttonsContainer.appendChild(arButton);

            card.appendChild(organizer);
            card.appendChild(eventName);
            card.appendChild(date);
            card.appendChild(buttonsContainer);

            container2.appendChild(card);
        }
    });
}

function handleAccept(invitation) {
    data = {"id": parseInt(invitation.meeting_id)}
    fetch("/acceptInvitation", {
        method: "POST",
        "Content-Type": "application/json",
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        location.reload();
    })
    .catch(error => console.error(error));
}

function handleReject(invitation) {
    data = {"id": parseInt(invitation.meeting_id)}
    fetch("/rejectInvitation", {
        method: "POST",
        "Content-Type": "application/json",
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        location.reload();
    })
    .catch(error => console.error(error));
}

