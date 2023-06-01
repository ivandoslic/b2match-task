var organizations = {};
var usersEvents = {};
var eventAttendees = {};
var usersMeetings = {};
var inviteesForEvent = {};
var selectedEvent = {};
var gatheredTimes = {};
var selectedTime = "";

function init() {
    getCurrentUserEvents();
}

function fillEvents() {
    const selectElement = document.getElementById('eventsAttending');
    let firstKey;
    for(let key in usersEvents) {
        firstKey = firstKey == null ? key : firstKey;
        const option = document.createElement('option');
        option.value = usersEvents[key].id;
        option.text = usersEvents[key].name;
        selectElement.appendChild(option);
    }
    selectElement.addEventListener('change', () => {
        getSelectedEventAttendees(selectElement.value);
    });
    getSelectedEventAttendees(firstKey);
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
        getCurrentUserMeetings();
    });
}

function generateUserList() {
    const userListDiv = document.getElementById('userList');
    userListDiv.innerHTML = '';

    const searchInput = document.getElementById('searchInput');
    const searchTerm = searchInput.value.toLowerCase();

    for(let key in eventAttendees) {
        if (eventAttendees[key].username.toLowerCase().includes(searchTerm)) {
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.value = eventAttendees[key].user_id;

            const label = document.createElement('label');
            label.textContent = eventAttendees[key].username;
            label.prepend(checkbox);

            userListDiv.appendChild(label);
            userListDiv.appendChild(document.createElement('br'));
        }
    }

    searchInput.addEventListener('input', () => generateUserList());
    
}

function getSelectedEventAttendees(eventId) {
    data = {"id": parseInt(eventId)}
    attendees = {}
    fetch("/getAttendees", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        eventAttendees = {};
        data.forEach(user => {
            eventAttendees[user.user_id] = user;
        })
    })
    .then(() => generateUserList())
    .catch(error => console.error(error));
}

function sendInvitations() {
    const selectedUsers = Array.from(document.querySelectorAll('input[type="checkbox"]:checked'))
        .map(checkbox => parseInt(checkbox.value));

    const duration = parseInt(document.getElementById('meetingDurationInput').value);
    const selectedEvent = usersEvents[document.getElementById('eventsAttending').value];
    console.log("SELECTED:", JSON.stringify(selectedEvent));

    var ids = [];
    var count = 0;
    
    for(let i in selectedUsers) {
        var tempBody = {
            "id": parseInt(selectedUsers[i])
        }
        ids[count] = tempBody;
        count++;
    }

    data = {
        "event": selectedEvent,
        "duration": parseInt(duration),
        "invited_users_ids": ids
    }

    console.log(JSON.stringify(data));

    fetch('/sendInvitations', {
        method: "POST",
        "Content-Type": "application/json",
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => console.log(data))
    .then(() => location.reload())
    .catch(error => console.error(error));
}

function fillMeetingOptions() {
    const meetingSelectElement = document.getElementById('meetingsOrganized');
    let firstKey;
    for(let key in usersMeetings) {
        firstKey = firstKey == null ? key : firstKey;
        const option = document.createElement('option');
        option.value = usersMeetings[key].meeting_id;
        option.text = usersEvents[usersMeetings[key].event_id].name + " (" + key + ")";
        meetingSelectElement.appendChild(option);
    }
    meetingSelectElement.addEventListener('change', () => {
        getInviteesForMeeting(meetingSelectElement.value);
    })
    getInviteesForMeeting(firstKey);
}

function getCurrentUserMeetings() {
    fetch("/getUsersMeetings", {
        method: "GET"
    })
    .then(response => response.json())
    .then(data => {
        usersMeetings = {};
        data.forEach(meeting => {
            usersMeetings[meeting.meeting_id] = meeting
        });
    })
    .then(() => fillMeetingOptions())
    .catch(error => console.error(error));
}

function getInviteesForMeeting(meetingId) {
    data = {"id": parseInt(meetingId)};
    fetch("/getInviteesForMeeting", {
        method: "POST",
        "Content-Type": "application/json",
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        inviteesForEvent = data;
    })
    .then(() => updateSelectedMeetingInfo(meetingId))
    .catch(error => console.error(error));
}

function updateSelectedMeetingInfo(meetingId) {
    const meetingIdLabel = document.getElementById('meetingIdLabel');
    const eventNameLabel = document.getElementById('eventNameLabel');
    const dateLabel = document.getElementById('dateLabel');
    const timeLabel = document.getElementById('timeLabel');
    const durationLabel = document.getElementById('durationLabel');
    const inviteeInfoDiv = document.getElementById('inviteeInfo');

    meetingIdLabel.innerHTML = "Meeting ID: " + usersMeetings[meetingId].meeting_id;
    eventNameLabel.innerHTML = "Event: " + usersEvents[usersMeetings[meetingId].event_id].name;
    dateLabel.innerHTML = "Date: " + usersMeetings[meetingId].date;
    timeLabel.innerHTML = usersMeetings[meetingId].time === "" ? "Time: TBD" : "Time: " + usersMeetings[meetingId].time;
    durationLabel.innerHTML = "Duration: " + usersMeetings[meetingId].duration + " minutes";
    inviteeInfoDiv.innerHTML = '';

    inviteesForEvent.forEach(inviteeAndInvitation => {
        const card = document.createElement("div");
        card.className = "card";

        const name = document.createElement("div");
        name.className = "name";
        name.textContent = inviteeAndInvitation.invitee.username;

        const email = document.createElement("div");
        email.className = "email";
        email.textContent = inviteeAndInvitation.invitee.email;

        const status = document.createElement("div");
        status.className = "status";
        status.textContent = `Status: ${inviteeAndInvitation.invitation.status}`;

        card.appendChild(name);
        card.appendChild(email);
        card.appendChild(status);

        inviteeInfoDiv.appendChild(card);
    });
    
    data = {"id": usersMeetings[meetingId].meeting_id}
    fetch("/getPossibleTimes", {
        method: "POST",
        "Content-Type": "application/json",
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        if(data.status === "unavailable") {
            setUnavailableMessage(data.time[0]);
        } else {
            if(usersMeetings[meetingId].time === "")
                setAvailableMessage(data.time);
            else
                showTimeAlreadyDetermined();
        }
    })
}

function showTimeAlreadyDetermined() {
    const schedulingInfoDiv = document.getElementById('timeSchedulingDiv');
    schedulingInfoDiv.innerHTML = '';
    const paragraf = document.createElement('p');
    paragraf.textContent = "This meeting has been scheduled!";

    schedulingInfoDiv.appendChild(paragraf);
}

function setUnavailableMessage(text) {
    const schedulingInfoDiv = document.getElementById('timeSchedulingDiv');
    schedulingInfoDiv.innerHTML = '';
    const paragraf = document.createElement('p');
    paragraf.textContent = text;

    schedulingInfoDiv.appendChild(paragraf);
}

function setAvailableMessage(times) {
    const schedulingInfoDiv = document.getElementById('timeSchedulingDiv');
    schedulingInfoDiv.innerHTML = '';
    const selectElement = document.createElement('select');
    selectElement.id = "timeSelect";
    count = 0;
    gatheredTimes = {};
    times.forEach(time => {
        const optionElement = document.createElement('option');
        optionElement.textContent = time;
        optionElement.value = count;
        gatheredTimes[count] = time;
        count++;
        selectElement.appendChild(optionElement);
    });
    selectedTime = times[0];
    schedulingInfoDiv.appendChild(selectElement);
    selectElement.addEventListener('change', () => {
            selectedTime = selectElement.value;
    });
    const scheduleButton = document.createElement('button');
    const brElement = document.createElement('br');
    scheduleButton.textContent = "Schedule";
    scheduleButton.onclick = () => {
        scheduleData = {
            "meeting_id": parseInt(document.getElementById('meetingsOrganized').value),
            "time": gatheredTimes[selectElement.value]
        }
        fetch("/scheduleMeetingTime", {
            method: "POST",
            "Content-Type": "application/json",
            body: JSON.stringify(scheduleData)
        })
        .then(response => response.json())
        .then(() => location.reload())
        .catch(error => console.error(error));
    }
    schedulingInfoDiv.appendChild(brElement);
    schedulingInfoDiv.appendChild(brElement);
    schedulingInfoDiv.appendChild(scheduleButton);
}