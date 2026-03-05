import { GraphDataProvider } from '@/contexts/GraphDataContext';
import { UIStateProvider } from '@/contexts/UIStateContext';
import { Layout } from '@/components/Layout/Layout';
import { ErrorBoundary } from '@/components/ErrorBoundary';

function App() {
  return (
    <div className="dark h-screen w-screen overflow-hidden">
      <ErrorBoundary>
        <GraphDataProvider>
          <UIStateProvider>
            <Layout />
          </UIStateProvider>
        </GraphDataProvider>
      </ErrorBoundary>
    </div>
  );
}

export default App;
