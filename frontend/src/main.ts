import { bootstrapApplication } from '@angular/platform-browser';
import { appConfig } from './app/app.config';
import { App } from './app/app';
import { 
  Chart, 
  CategoryScale, 
  LinearScale, 
  PointElement, 
  LineElement, 
  LineController,
  BarElement,
  BarController,
  ArcElement,
  DoughnutController,
  Title, 
  Tooltip, 
  Legend, 
  Filler,
  TimeScale,
  TimeSeriesScale
} from 'chart.js';

// Register Chart.js components
Chart.register(
  CategoryScale,
  LinearScale,
  TimeScale,
  TimeSeriesScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  LineController,
  BarController,
  DoughnutController,
  Title,
  Tooltip,
  Legend,
  Filler
);

bootstrapApplication(App, appConfig)
  .catch((err) => console.error(err));
