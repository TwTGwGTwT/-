document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('stats-form');
    const catCanvas = document.getElementById('by-category-chart');
    const monthCanvas = document.getElementById('by-month-chart');

    // Год по умолчанию
    form.elements['year'].value = new Date().getFullYear();

    let catChart = null;
    let monthChart = null;

    const monthNames = {
        '01': 'январь', '02': 'февраль', '03': 'март',
        '04': 'апрель', '05': 'май', '06': 'июнь',
        '07': 'июль', '08': 'август', '09': 'сентябрь',
        '10': 'октябрь', '11': 'ноябрь', '12': 'декабрь'
    };

    function prettyMonth(m) {
        const parts = m.split('-');
        return (monthNames[parts[1]] || parts[1]) + ' ' + parts[0];
    }

    function buildCategoryChart(data) {
        if (catChart) catChart.destroy();
        catChart = new Chart(catCanvas, {
            type: 'pie',
            data: {
                labels: data.map(s => s.CategoryName),
                datasets: [{
                    data: data.map(s => s.Total)
                }]
            }
        });
    }

    function buildMonthChart(data) {
        if (monthChart) monthChart.destroy();
        monthChart = new Chart(monthCanvas, {
            type: 'bar',
            data: {
                labels: data.map(s => prettyMonth(s.Month)),
                datasets: [{
                    label: 'Сумма трат',
                    data: data.map(s => s.Total)
                }]
            },
            options: {
                scales: {
                    y: { beginAtZero: true }
                }
            }
        });
    }

    async function loadStats() {
        const values = new FormData(form);

        // Параметры для категорий
        const catParams = new URLSearchParams();
        if (values.get('from')) catParams.set('from', values.get('from'));
        if (values.get('to')) catParams.set('to', values.get('to'));

        // Параметры для месяцев
        const monthParams = new URLSearchParams();
        if (values.get('year')) monthParams.set('year', values.get('year'));

        const catResp = await fetch('/api/stats/by-category?' + catParams.toString());
        if (!catResp.ok) {
            console.error('Не удалось загрузить статистику по категориям');
            return;
        }
        const catData = await catResp.json();
        buildCategoryChart(catData);

        const monthResp = await fetch('/api/stats/by-month?' + monthParams.toString());
        if (!monthResp.ok) {
            console.error('Не удалось загрузить статистику по месяцам');
            return;
        }
        const monthData = await monthResp.json();
        buildMonthChart(monthData);
    }

    form.addEventListener('submit', function (e) {
        e.preventDefault();
        loadStats();
    });

    document.getElementById('reset-btn').addEventListener('click', function () {
        form.reset();
        form.elements['year'].value = new Date().getFullYear();
        loadStats();
    });

    loadStats();
});