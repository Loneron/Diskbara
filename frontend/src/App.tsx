function App() {
  return (
    <main className="container">
      <div className="hero">
        <div className="hero-content">
          <div className="badge">Version 0.1.0</div>
          <h1 className="title">
            <span className="title-accent">Diskbara</span>
          </h1>
          <p className="subtitle">
            Держи свой диск на чиле 🦫
          </p>
          <section>
            <h2>Disk Space Analyzer</h2>
            <p>
              Узнай, что занимает место на твоем компуктере.
            </p>
            <button className="btn" disabled>
              Выбери папку
            </button>
          </section>
        </div>

        <div className="hero-illustration">
          <img
            src="/images/diskbara-mascot.png"
            alt="Капибара Diskbara отдыхает рядом с жёстким диском"
            className="mascot"
          />
        </div>

      </div>
    </main>
  );
}

export default App;
