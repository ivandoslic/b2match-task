var organizations = {};
var usersEvents = {};

function init() {
    fillOptions();
}

function fillEvents() {
    const eventList = document.getElementById('eventList');
    const attentingList = document.getElementById('attendingList');
    fetch('/getEvents')
    .then(response => response.json())
    .then(data => {
        data.forEach(event => {
            if(usersEvents[event.id] == null)
                eventList.appendChild(createCard(event, false))
            else
                attentingList.appendChild(createCard(event, true))
        });
    })
    .catch(error => {
        console.error("Error fetching events:", error);
    });
}

function fillOptions() {
    const formSelect = document.getElementById('formEventOrganizator');
    fetch('/getOrganizations')
    .then(response => response.json())
    .then(data => {
        data.forEach(org => {
            const option = document.createElement('option');
            option.value = org.id;
            option.text = org.organization_name;
            organizations[org.id] = org;
            formSelect.appendChild(option);
        });
    })
    .then(() => {
        var authMeta = document.querySelector('meta[name="authStatus"');
        if(authMeta.getAttribute('content').localeCompare("true") == 0) {
            getCurrentUserEvents();
        }
    })
    .catch(error => {
        console.error("Error fetching organizations:", error);
    });
}

function createEventInDatabase() {
    const nameInput = document.getElementById('formEventName');
    const dateInput = document.getElementById('formEventDate');
    const startInput = document.getElementById('formEventStart');
    const endInput = document.getElementById('formEventEnd');
    const orgInput = document.getElementById('formEventOrganizator');

    let data = {
        "id": 0,
        "name": nameInput.value,
        "date": dateInput.value,
        "organizator": orgInput.value,
        "start_time": startInput.value,
        "end_time": endInput.value
    }  

    fetch("/addEvent", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(data)
    })
    .then(response => {
        console.log(response.status);
        console.log(response.statusText);
        console.log(response.headers);

        return response.json();
    })
    .then((data) => {
        console.log(data);
        location.reload();
    })
    .catch(error => {
        console.error(error);
    });
}

function createCard(event, attending) {
    var eventID = event.id;
    var eventName = event.name;
    var eventOrganizer = event.organizator;
    var date = event.date;
    var startTime = event.start_time;
    var endTime = event.end_time;

    const cardContainer = document.createElement('div');
    cardContainer.classList.add('card');
  
    const cardHeader = document.createElement('div');
    cardHeader.classList.add('card-header');
  
    const eventNameElement = document.createElement('h3');
    eventNameElement.classList.add('event-name');
    eventNameElement.textContent = eventName;
  
    const eventOrganizerElement = document.createElement('p');
    eventOrganizerElement.classList.add('event-organizer');
    eventOrganizerElement.textContent = organizations[eventOrganizer].organization_name;
  
    cardHeader.appendChild(eventNameElement);
    cardHeader.appendChild(eventOrganizerElement);
  
    const cardBody = document.createElement('div');
    cardBody.classList.add('card-body');
  
    const eventIDElement = document.createElement('p');
    eventIDElement.classList.add('event-id');
    eventIDElement.textContent = `Event ID: ${eventID}`;
  
    const dateElement = document.createElement('p');
    dateElement.classList.add('event-date');
    dateElement.textContent = `Date: ${date}`;
  
    const startTimeElement = document.createElement('p');
    startTimeElement.classList.add('event-start-time');
    startTimeElement.textContent = `Start Time: ${startTime}`;
  
    const endTimeElement = document.createElement('p');
    endTimeElement.classList.add('event-end-time');
    endTimeElement.textContent = `End Time: ${endTime}`;
  
    cardBody.appendChild(eventIDElement);
    cardBody.appendChild(dateElement);
    cardBody.appendChild(startTimeElement);
    cardBody.appendChild(endTimeElement);
  
    const cardFooter = document.createElement('div');
    cardFooter.classList.add('card-footer');
  
    const joinButton = document.createElement('button');
    joinButton.classList.add(!attending ? 'join-button' : 'cancel-button');
    joinButton.textContent = !attending ? 'Join' : 'Leave';

    joinButton.onclick = function() {
        if(!attending)
            joinEvent(eventID);
        else
            leaveEvent(eventID);
    }
  
    cardFooter.appendChild(joinButton);
  
    cardContainer.appendChild(cardHeader);
    cardContainer.appendChild(cardBody);
    cardContainer.appendChild(cardFooter);
  
    return cardContainer;
}

function joinEvent(eventId) {
    data = {
        "id": eventId
    }

    fetch("/joinEvent", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        console.log(data);
        location.reload();
    })
    .catch(error => {
        console.error(error);
    });
}

function leaveEvent(eventId) {
    data = {
        "id": eventId
    }

    fetch("/leaveEvent", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        console.log(data);
        location.reload();
    })
    .catch(error => {
        console.error(error)
    });
}

function getCurrentUserEvents() {
    fetch("/getParticipations", {
        method: "GET",
    })
    .then(response => response.json())
    .then(data => {
        data.forEach(event => {
            usersEvents[event.id] = event;
        })
    })
    .then(() => {
        fillEvents();
    });
}