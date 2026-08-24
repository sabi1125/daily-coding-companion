import { env } from "@/lib/env"
import { Button } from "./ui/button"
import { FaGoogle } from "react-icons/fa"

function LoginState() {
  return (
    < div className="flex flex-col justify-center items-center w-auto h-screen gap-6" >

      {/* logo */}

      <div className="flex size-11 items-center justify-center rounded-xl bg-foreground font-mono text-lg font-semibold text-background">
        {">_"}
      </div>

      {/* header */}

      <div className="flex flex-col justify-center items-center">
        <h1 className="font-semibold text-2xl pb-3">Coding Companion</h1>
        <p className="text-sm text-text-faint pb-2">Your daily coding problems, with hints when you are stuck.</p>
      </div>

      {/* card */}

      <div className="border border-border-faint rounded-lg px-8 py-6 flex flex-col justify-center items-center gap-5 origin-center text-center">
        <Button className="px-19 py-5 gap-3 cursor-pointer" onClick={() => { window.location.assign(`${env.apiBaseUrl}/auth/google`) }}>
          <span className="flex size-5 items-center justify-center rounded-full bg-background border border-border-faint">
            <FaGoogle className="size-2.5 text-foreground font-bold" />
          </span>
          <p>Continue with Google</p>
        </Button>
        <p className="text-xs text-text-faint w-85">Coding Companion reads only your daily problem emails. It never sends mail or reads anything else.</p>
      </div>
      <p className="text-xs text-text-faint">Access is limited to approved users during testing.</p>
    </div >
  )
}

export default LoginState
