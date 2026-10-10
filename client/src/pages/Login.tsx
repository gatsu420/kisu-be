import { useSearchParams } from 'react-router-dom'
import { redirectToLogin } from '../lib/api'
import styles from './Login.module.css'

export default function Login() {
  const [searchParams] = useSearchParams()
  const errorMsg = searchParams.get('error')

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h1 className={styles.title}>kisu</h1>
        <p className={styles.subtitle}>Query your data with AI</p>
        {errorMsg && (
          <p className={styles.error}>
            {errorMsg === 'backend_unavailable'
              ? 'Backend is not available. Make sure kisu-be is running on :8080.'
              : errorMsg}
          </p>
        )}
        <button className={styles.button} onClick={redirectToLogin}>
          Sign in with Google
        </button>
      </div>
    </div>
  )
}
