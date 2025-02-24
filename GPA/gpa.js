document.addEventListener("DOMContentLoaded", function () {
    const addModuleBtn = document.getElementById("add-module");
    const calculateBtn = document.getElementById("calculate-gpa");
    const tableBody = document.getElementById("modules-body");
    const resultBox = document.getElementById("gpa-result");

    // Function to add a new module row
    addModuleBtn.addEventListener("click", function () {
        let newRow = document.createElement("tr");
        newRow.innerHTML = `
            <td><input type="text" class="module-name"></td>
            <td><input type="number" class="credit-unit" min="1"></td>
            <td><input type="text" class="grade"></td>
        `;
        tableBody.appendChild(newRow);
    });

    // Function to calculate GPA
    calculateBtn.addEventListener("click", function () {
        let rows = document.querySelectorAll("#modules-body tr");
        let totalCreditUnits = 0;
        let totalGradePoints = 0;
        let valid = true;

        rows.forEach(row => {
            let creditUnit = row.querySelector(".credit-unit").value;
            let grade = row.querySelector(".grade").value.toUpperCase(); 

            if (creditUnit === "" || grade === "") {
                valid = false;
                return;
            }

            creditUnit = parseInt(creditUnit);
            let gradePoint = getGradePoint(grade);

            if (isNaN(creditUnit) || gradePoint === null) {
                valid = false;
                return;
            }

            totalCreditUnits += creditUnit;
            totalGradePoints += creditUnit * gradePoint;
        });

        if (!valid || totalCreditUnits === 0) {
            resultBox.innerHTML = `<p style="color:red;">Invalid input. Please check Credit Units and Grades.</p>`;
            return;
        }

        let gpa = (totalGradePoints / totalCreditUnits).toFixed(2);
        
        // Display result in GPA box
        resultBox.innerHTML = `
            <div class="result-container">
                <h3>Results</h3>
                <p>GPA: <strong>${gpa}</strong></p>
            </div>
        `;
    });

    // Function to convert letter grade to grade points
    function getGradePoint(grade) {
        const gradeScale = {
            "A": 4.0,
            "B+": 3.5,
            "B": 3.0,
            "C+": 2.5,
            "C": 2.0,
            "D+": 1.5,
            "D": 1.0,
            "F": 0.0
        };
        return gradeScale[grade] || null;
    }
});
