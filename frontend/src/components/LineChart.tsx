import { useRef, useEffect } from 'react';

interface DataPoint {
  timestamp: string;
  value: number;
}

interface LineChartProps {
  data: DataPoint[];
  label: string;
  color?: string;
  height?: number;
  yAxisLabel?: string;
}

export function LineChart({ data, label, color = '#2563eb', height = 250, yAxisLabel }: LineChartProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || data.length === 0) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = (height || 250) * dpr;
    ctx.scale(dpr, dpr);

    const width = rect.width;
    const chartHeight = height || 250;
    const padding = { top: 20, right: 30, bottom: 40, left: 50 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeightInner = chartHeight - padding.top - padding.bottom;

    if (chartWidth <= 0 || chartHeightInner <= 0) return;

    const values = data.map((d) => d.value);
    const minValue = Math.min(...values, 0);
    const maxValue = Math.max(...values);
    const valueRange = maxValue - minValue || 1;

    ctx.clearRect(0, 0, width, chartHeight);

    // Grid lines
    ctx.strokeStyle = '#e5e7eb';
    ctx.lineWidth = 1;
    const gridLines = 4;
    for (let i = 0; i <= gridLines; i++) {
      const y = padding.top + (chartHeightInner / gridLines) * i;
      ctx.beginPath();
      ctx.moveTo(padding.left, y);
      ctx.lineTo(width - padding.right, y);
      ctx.stroke();

      // Y-axis labels
      const val = maxValue - (valueRange / gridLines) * i;
      ctx.fillStyle = '#9ca3af';
      ctx.font = '11px system-ui';
      ctx.textAlign = 'right';
      ctx.fillText(val.toFixed(1), padding.left - 8, y + 4);
    }

    // Y-axis label
    if (yAxisLabel) {
      ctx.save();
      ctx.translate(15, padding.top + chartHeightInner / 2);
      ctx.rotate(-Math.PI / 2);
      ctx.fillStyle = '#9ca3af';
      ctx.font = '11px system-ui';
      ctx.textAlign = 'center';
      ctx.fillText(yAxisLabel, 0, 0);
      ctx.restore();
    }

    // X-axis line
    ctx.strokeStyle = '#d1d5db';
    ctx.beginPath();
    ctx.moveTo(padding.left, padding.top + chartHeightInner);
    ctx.lineTo(width - padding.right, padding.top + chartHeightInner);
    ctx.stroke();

    // X-axis labels (timestamps)
    ctx.fillStyle = '#9ca3af';
    ctx.font = '11px system-ui';
    ctx.textAlign = 'center';
    const labelCount = Math.min(data.length, 6);
    for (let i = 0; i < labelCount; i++) {
      const idx = Math.floor((data.length - 1) * (i / (labelCount - 1 || 1)));
      const x = padding.left + (chartWidth / (labelCount - 1 || 1)) * i;
      const date = new Date(data[idx].timestamp);
      const timeStr = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      ctx.fillText(timeStr, x, padding.top + chartHeightInner + 20);
    }

    // Draw line
    if (data.length > 1) {
      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.lineCap = 'round';
      ctx.lineJoin = 'round';
      ctx.beginPath();

      data.forEach((point, i) => {
        const x = padding.left + (chartWidth / (data.length - 1)) * i;
        const y = padding.top + chartHeightInner - ((point.value - minValue) / valueRange) * chartHeightInner;
        if (i === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });

      ctx.stroke();

      // Draw points
      ctx.fillStyle = color;
      data.forEach((point, i) => {
        const x = padding.left + (chartWidth / (data.length - 1)) * i;
        const y = padding.top + chartHeightInner - ((point.value - minValue) / valueRange) * chartHeightInner;
        ctx.beginPath();
        ctx.arc(x, y, 3, 0, Math.PI * 2);
        ctx.fill();
      });
    }

    // Legend
    ctx.fillStyle = color;
    ctx.font = '12px system-ui';
    ctx.textAlign = 'left';
    ctx.fillText(label, padding.left, padding.top - 5);
  }, [data, label, color, height, yAxisLabel]);

  return (
    <div className="chart-container">
      <canvas ref={canvasRef} style={{ width: '100%', height: '100%' }} aria-label={`${label} chart`} />
    </div>
  );
}