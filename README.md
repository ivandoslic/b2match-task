# b2match Task Solution

This application requires a few dependencies to be installed and configured before you can run it successfully. Please follow the instructions below to set up your environment.

## Prerequisites

Before running this application, make sure you have the following prerequisites installed:

- MySQL Server
- MySQL Workbench
- Go

## Installation

1. Clone the repository to your local machine.
```bash
git clone https://github.com/ivandoslic/b2match-task.git
```

2. Open a command prompt or shell and navigate to the project directory.
```bash
cd b2match-task
```

3. Create the database schema by running the following command as root:
```bash
mysql -u root -p < db/init.sql
```
Enter your root password when prompted.

## Running the Application

To run the application, follow these steps:

1. Open a command prompt or shell inside the project directory.

2. Execute the following command:
```bash
go run ./main.go
```

The application will now start running, and you can access it by opening a web browser and navigating to http://localhost:8080/

## Additional Notes

Make sure that in main.go you put your password for root!

## Support

If you encounter any issues or have any questions, please feel free to contact at doslicivan03@gmail.com