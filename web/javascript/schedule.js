var wholeSchedule = {}
var events = {}

function init() {
    fetch("/getEvents", {
        method: "GET"
    })
    .then(response => response.json())
    .then(data => {
        data.forEach(event => {
            events[event.id] = event;
        })
    })
    .then(() => getWholeSchedule())
    .catch(error => console.error(error));
}

function getWholeSchedule() {
    fetch("/getUserSchedule", {
        method: "GET"
    })
    .then(response => response.json())
    .then(data => {
        data.schedule_entries.forEach(entry => {
            wholeSchedule[entry.meeting_id] = entry
        })
    })
    .then(() => setupSelectElement())
    .catch(error => console.error(error));
}

function setupSelectElement() {
    const selectElement = document.getElementById('eventOptions');
    for(let key in events) {
        if(!userAttends(events[key].id)) continue;
        const optionElement = document.createElement('option');
        optionElement.textContent = events[key].name;
        optionElement.value = key;
        console.log(key, events[key])

        selectElement.appendChild(optionElement);
    }
    selectElement.addEventListener('change', () => {
        displayEventDetails(events[parseInt(selectElement.value)]);
        displayUserSchedule(getScheduleFor(parseInt(selectElement.value)));
    })
    displayEventDetails(events[parseInt(selectElement.value)]);
    displayUserSchedule(getScheduleFor(parseInt(selectElement.value)));
}

function getScheduleFor(eventId) {
    entries = {}
    for(let key in wholeSchedule) {
        if(wholeSchedule[key].event_id === eventId) entries[wholeSchedule[key].meeting_id] = wholeSchedule[key];
    }
    return entries;
}

function userAttends(eventId) {
    for(let key in wholeSchedule)
        if(wholeSchedule[key].event_id === eventId) return true;
    return false;
}

// Function to display the event details
function displayEventDetails(event) {
    console.log("Displaying new event!", event);
    const eventNameElement = document.getElementById('event-name');
    const eventDateElement = document.getElementById('event-date');
    const eventTimeElement = document.getElementById('event-time');

    eventNameElement.textContent = event.name;
    eventDateElement.textContent = event.date;
    eventTimeElement.textContent = `${event.start_time} - ${event.end_time}`;
}

// Function to display the user's schedule for the event
function displayUserSchedule(scheduleEntries) {
    const calendarTimeColumn = document.getElementById('timeColumn');
    const calendarScheduleColumn = document.getElementById('scheduleColumn');

    // Clear the schedule columns
    calendarTimeColumn.innerHTML = '';
    calendarScheduleColumn.innerHTML = '';

    // Iterate over the schedule entries and create time and schedule blocks
    for(let key in scheduleEntries) {
        const timeBlock = document.createElement('div');
        timeBlock.classList.add('calendar-time-block');
        timeBlock.textContent = scheduleEntries[key].start_time;

        const scheduleBlock = document.createElement('div');
        scheduleBlock.classList.add('calendar-schedule-block');
        scheduleBlock.innerHTML = `<span>Meeting ID:</span> ${key}<br>
                                    <span>Start Time:</span> ${scheduleEntries[key].start_time}<br>
                                    <span>End Time:</span> ${scheduleEntries[key].end_time}<p>   <p>`;

        calendarScheduleColumn.appendChild(timeBlock);
        calendarScheduleColumn.appendChild(scheduleBlock);
    }
}