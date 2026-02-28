(function () {
  'use strict';

  // Render fan chart if stats are present on the results page.
  function renderFanChart() {
    const canvas = document.getElementById('fanChart');
    if (!canvas || !window.DRIFT_STATS) return;

    const stats = window.DRIFT_STATS;
    // Build a simple bar chart of percentile terminal values as a proxy
    // (real fan chart requires path data; this shows the percentile distribution).
    new Chart(canvas, {
      type: 'bar',
      data: {
        labels: ['p5', 'p25', 'p50', 'p75', 'p95'],
        datasets: [{
          label: 'Terminal Portfolio Value ($)',
          data: [stats.P5, stats.P25, stats.P50, stats.P75, stats.P95],
          backgroundColor: [
            'rgba(99,102,241,0.3)',
            'rgba(99,102,241,0.5)',
            'rgba(99,102,241,0.9)',
            'rgba(99,102,241,0.5)',
            'rgba(99,102,241,0.3)',
          ],
          borderColor: 'rgba(99,102,241,1)',
          borderWidth: 1,
          borderRadius: 4,
        }],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (ctx) => '$' + ctx.parsed.y.toLocaleString(undefined, {maximumFractionDigits: 0}),
            },
          },
        },
        scales: {
          y: {
            ticks: {
              color: '#64748b',
              callback: (v) => '$' + v.toLocaleString(undefined, {maximumFractionDigits: 0}),
            },
            grid: { color: '#2a2d3a' },
          },
          x: {
            ticks: { color: '#64748b' },
            grid: { display: false },
          },
        },
      },
    });
  }

  document.addEventListener('DOMContentLoaded', renderFanChart);
})();
