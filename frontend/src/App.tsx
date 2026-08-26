import { Toaster } from 'sonner'
import Auth from './features/auth/Auth.tsx'

function App() {
  return (
    <>
      <Toaster position="bottom-center" toastOptions={{
        classNames: {
          toast: "!bg-foreground !text-background !border-transparent !w-fit !inset-x-0 !mx-auto"
        }
      }}
      />
      <Auth />
    </>
  )
}

export default App
