import { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Button } from './ui/button';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary yakaladi:', error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center min-h-screen bg-zinc-950 p-8">
          <div className="rounded-xl border border-zinc-800 bg-zinc-900 p-8 max-w-md w-full text-center">
            <div className="w-14 h-14 bg-red-500/15 rounded-full flex items-center justify-center mx-auto mb-4">
              <AlertTriangle className="w-7 h-7 text-red-400" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">Beklenmeyen Hata</h2>
            <p className="text-zinc-400 text-sm mb-4">
              Bir seyler yanlis gitti. Lutfen tekrar deneyin.
            </p>
            {this.state.error && (
              <div className="bg-zinc-800 rounded-lg p-3 mb-6 text-left">
                <p className="text-xs text-zinc-500 font-mono break-all">
                  {this.state.error.message}
                </p>
              </div>
            )}
            <Button onClick={this.handleReset}>
              <RefreshCw className="w-4 h-4" />
              Tekrar Dene
            </Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
