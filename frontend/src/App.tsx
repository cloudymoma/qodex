import { GraphDataProvider } from '@/contexts/GraphDataContext';
import { UIStateProvider } from '@/contexts/UIStateContext';
import { Layout } from '@/components/Layout/Layout';
import { ErrorBoundary } from '@/components/ErrorBoundary';
import { AccessGate } from '@/components/AccessGate';

function App() {
  return (
    <div className="dark h-screen w-screen overflow-hidden">
      <ErrorBoundary>
        <AccessGate>
          <GraphDataProvider>
            <UIStateProvider>
              <Layout />
            </UIStateProvider>
          </GraphDataProvider>
        </AccessGate>
      </ErrorBoundary>
    </div>
  );
}

export default App;
