<template>
  <div class="home-page">
    <div class="hero">
      <h1>Event <span>Processor</span></h1>
      <p>Quality Gateway for Event Processing</p>
    </div>
    
    <div class="features">
      <div class="feature-card">
        <h3>Events</h3>
        <p>View and manage GitHub events and tasks</p>
        <router-link to="/events" class="btn btn-primary">View Events</router-link>
      </div>

      <div class="feature-card">
        <h3>Resources</h3>
        <p>Manage executable resources</p>
        <router-link to="/resources" class="btn btn-primary">View Resources</router-link>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Home'
}
</script>

<style scoped>
.home-page {
  text-align: center;
}

.hero {
  padding: 4rem 0;
  margin-bottom: 3rem;
  position: relative;
}

.hero::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 600px;
  height: 300px;
  background: radial-gradient(ellipse at center, rgba(212, 175, 55, 0.1) 0%, transparent 70%);
  pointer-events: none;
}

.hero h1 {
  font-size: 3rem;
  font-weight: 700;
  color: var(--primary);
  margin-bottom: 0.75rem;
  letter-spacing: -0.05em;
  position: relative;
}

.hero h1 span {
  color: var(--accent);
}

.hero p {
  font-size: 1.25rem;
  color: var(--text-secondary);
  font-weight: 400;
}

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 2rem;
  max-width: 900px;
  margin: 0 auto;
}

.feature-card {
  background: var(--bg-card);
  border-radius: var(--radius);
  padding: 2.5rem 2rem;
  box-shadow: var(--shadow);
  border: 1px solid var(--border);
  text-align: center;
  transition: all 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-color: var(--accent);
}

.feature-card::before {
  content: '';
  display: block;
  width: 48px;
  height: 48px;
  margin: 0 auto 1.5rem;
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-hover) 100%);
  border-radius: 12px;
  opacity: 0.15;
}

.feature-card:nth-child(1)::before {
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23D4AF37' stroke-width='2'%3e%3cpath stroke-linecap='round' stroke-linejoin='round' d='M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' /%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: center;
  background-size: 24px;
  background-color: rgba(212, 175, 55, 0.1);
}

.feature-card:nth-child(2)::before {
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23D4AF37' stroke-width='2'%3e%3cpath stroke-linecap='round' stroke-linejoin='round' d='M5 12h14m-12-8v8m16-8v8M4 7h16M4 12h16M4 17h16' /%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: center;
  background-size: 24px;
  background-color: rgba(212, 175, 55, 0.1);
}

.feature-card h3 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--primary);
}

.feature-card p {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
  font-size: 0.9375rem;
  line-height: 1.6;
}

.feature-card .btn {
  width: 100%;
  justify-content: center;
}

@media (max-width: 768px) {
  .hero {
    padding: 2.5rem 0;
    margin-bottom: 2rem;
  }

  .hero h1 {
    font-size: 2rem;
  }

  .hero p {
    font-size: 1rem;
  }

  .features {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .feature-card {
    padding: 2rem 1.5rem;
  }
}
</style>
