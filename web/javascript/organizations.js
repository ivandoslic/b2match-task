function createNewOrganization() {
    let data = {
        "id": "0",
        "organization_name": document.getElementById('formOrganizationName').value
    }

    console.log(JSON.stringify(data))

    fetch("/addOrganization", {
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
    .then(data => {
        console.log(data);
        document.getElementById('formOrganizationName').value = ""
        location.reload();
    })
    .catch(error => {
        console.error(error);
    });
}

function getOrganizations() {
    fetch('/getOrganizations')
    .then(response => response.json())
    .then(data => {
        const organizationsList = document.getElementById('listOfOrganizations');
        console.log(data);
        data.forEach(org => {
            const listItem = document.createElement('li');
            listItem.textContent = "id:" + org.id + " name: " + org.organization_name;
            organizationsList.appendChild(listItem);
        });
        document.getElementById('jsonDisplayField').innerHTML = JSON.stringify(data);
    })
    .catch(error => {
        console.error("Error fetching organizations:", error);
    })
}
