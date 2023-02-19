const fetchStatus = async (jobNamespace, jobName) => {
    const statusAPI = new URL(window.location.href)
    statusAPI.pathname = '/api/jobs/status'
    statusAPI.searchParams.set('namespace', jobNamespace)
    statusAPI.searchParams.set('name', jobName)

    const response = await fetch(statusAPI)
    if (!response.ok) {
        throw new Error(`${response.status} ${response.statusText}`)
    }
    return await response.json()
}

const refreshStatus = async () => {
    const params = new URLSearchParams(window.location.search)
    const jobNamespace = params.get('namespace')
    const jobName = params.get('name')
    document.getElementById('job-namespace').innerText = jobNamespace
    document.getElementById('job-name').innerText = jobName

    try {
        const jobStatus = await fetchStatus(jobNamespace, jobName)
        document.getElementById('response').innerText = JSON.stringify(jobStatus, undefined, 2)
    } catch (e) {
        document.getElementById('errors').innerText = String(e)
        return
    }
    document.getElementById('errors').innerText = ''
}

window.setInterval(refreshStatus, 3000)
