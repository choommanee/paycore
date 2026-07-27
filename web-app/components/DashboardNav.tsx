import LogoutButton from "@/components/LogoutButton";

export default function DashboardNav({ active }: { active?: string }) {
  const items = [
    { href: "/dashboard", label: "หน้าหลัก" },
    { href: "/transactions", label: "ธุรกรรม" },
    { href: "/links", label: "ลิงก์ชำระเงิน" },
    { href: "/settings", label: "ตั้งค่า" },
  ];
  return (
    <header className="flex items-center justify-between mb-8">
      <nav className="flex items-center gap-4">
        <span className="text-xl font-semibold">PayCore</span>
        {items.map((it) => (
          <a
            key={it.href}
            href={it.href}
            className={
              "text-sm hover:text-paycore-text " +
              (active === it.href ? "text-paycore-text font-medium" : "text-paycore-muted")
            }
          >
            {it.label}
          </a>
        ))}
      </nav>
      <LogoutButton />
    </header>
  );
}
